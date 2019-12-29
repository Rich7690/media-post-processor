package worker

import (
	"fmt"
	"github.com/gocraft/work"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
	"log"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/utils"
	"media-web/internal/web"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

type Webhook struct {
	EventType string `json:"eventType"`
}

type WorkScheduler struct {
	EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
}

var worker = work.NewEnqueuer(constants.JobQueueNamespace, &storage.RedisPool)

var Enqueuer = WorkScheduler{
	EnqueueUnique: worker.EnqueueUnique,
}

func (c *Webhook) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Starting job: ", job.ID)
	return next()
}

func getMovieFilePath(id int64) (string, error) {
	movie, err := web.LookupMovie(id)
	if err != nil {
		return "", err
	}
	if movie != nil {
		fmt.Println("Got movie: ", movie.Title)
		return movie.Path + "/" + movie.MovieFile.RelativePath, nil
	} else {
		fmt.Println("Could not find movie")
	}

	return "", nil
}

func getEpisodeFilePath(id int64) (string, int, error) {
	episodeFile, err := web.LookupTVEpisode(id)
	if err != nil {
		return "", -1, err
	}
	if episodeFile != nil {
		fmt.Println("Got episodeFile: ", episodeFile.RelativePath)
		return episodeFile.Path, episodeFile.SeriesID, nil
	} else {
		fmt.Println("Could not find episodeFile")
	}
	return "", -1, nil
}

func (c *Webhook) UpdateTVShow(job *work.Job) error {

	seriesId := job.ArgInt64(constants.SeriesIdKey)

	cmd, err := web.RescanSeries(seriesId)

	if err != nil {
		fmt.Println("Error rescanning: ", err)
	}

	for count := 0; count < 5; count++ {
		result, err := web.CheckSonarrCommand(cmd.ID)

		if err == nil {
			if strings.Index(result.State, "complete") != -1 {
				fmt.Println("Rescan complete for: ", cmd.ID)
				return nil
			} else {
				fmt.Println("Rescan not complete yet for: ", cmd.ID)
			}
		} else {
			fmt.Println("Error: ", err)
		}

		time.Sleep(time.Second * 15)
	}

	return err
}

func (c *Webhook) UpdateMovie(job *work.Job) (error) {
	movieId := job.ArgInt64(constants.MovieIdKey)

	cmd, err := web.RescanMovie(movieId)

	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	for count := 0; count < 5; count++ {
		result, err := web.CheckRadarrCommand(cmd.ID)

		if err == nil {
			if strings.Index(result.State, "complete") != -1 {
				fmt.Println("Rescan complete for: ", cmd.ID)
				return nil
			} else {
				fmt.Println("Rescan not complete yet for: ", cmd.ID)
			}
		} else {
			fmt.Println("Error: ", err)
		}

		time.Sleep(time.Second * 15)
	}

	return nil
}

func (c *Webhook) DoTranscode(job *work.Job) error {
	transcodeType := constants.TranscodeType(job.ArgString(constants.TranscodeTypeKey))

	var inputFilePath = ""
	var id int64 = -1
	var err error = nil
	var seriesId = -1
	switch transcodeType {
	case constants.TV:
		id = job.ArgInt64(constants.EpisodeFileIdKey)
		inputFilePath, seriesId, err = getEpisodeFilePath(id)
	case constants.Movie:
		id = job.ArgInt64(constants.MovieIdKey)
		inputFilePath, err = getMovieFilePath(id)
	default:
		fmt.Println("Unknown transcodeType: ", transcodeType)
		return nil
	}

	if err != nil {
		log.Println("Error: " + err.Error())
		return err
	}

	if inputFilePath != "" {
		log.Println("Working on transcode at path: ", inputFilePath)
	} else {
		log.Println("Could not get input file path")
		return nil
	}
	// Create new instance of transcoder
	trans := new(transcoder.Transcoder)
	trans.SetConfiguration(ffmpeg.Configuration{
		FfmpegBin:  config.GetFfmpegPath(),
		FfprobeBin: config.GetFfprobePath(),
	})

	if !utils.FileExists(inputFilePath) {
		log.Println("Could not find file at path: ", inputFilePath)
		return nil
	}
	// Initialize transcoder passing the input file path and output file path
	ext := filepath.Ext(inputFilePath)

	log.Println("Current extension: ", ext)
	if ext == ".mp4" {
		log.Println("File is already mp4 extension. Skipping...")
		return nil
	}

	fileName := filepath.Base(inputFilePath)
	baseDir := filepath.Dir(inputFilePath)
	newPath := baseDir + "/" + strings.Replace(fileName, ext, ".mp4", 1)
	log.Println("Transcoding to path: ", newPath)
	err = trans.Initialize(inputFilePath, newPath)

	if err != nil {
		fmt.Println("Error initializing transcode: ", err)
		return err
	}

	fmt.Println("Transcoding: ", trans.MediaFile().InputPath())

	trans.MediaFile().SetPreset("veryfast")
	trans.MediaFile().SetOutputFormat("mp4")
	trans.MediaFile().SetVideoCodec("libx264")
	trans.MediaFile().SetQuality(23)
	trans.MediaFile().SetTune("film")

	log.Println("Running ffmpeg command: \"", trans.MediaFile().ToStrCommand(), "\"")

	// Start transcoder process with progress checking
	done := trans.Run(true)

	// Returns a channel to get the transcoding progress
	progress := trans.Output()

	// Example of printing transcoding progress
	for msg := range progress {
		message := "Transcoding: " + inputFilePath + " -> " + fmt.Sprint(msg)
		fmt.Println(message)
		job.Checkin(message)
	}

	// This channel is used to wait for the transcoding process to end
	err = <-done

	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Println("Deleting old file")

	err = os.Remove(inputFilePath)

	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Done transcoding: ", newPath)

	if transcodeType == constants.TV {
		updateJob, err := Enqueuer.EnqueueUnique("update-sonarr", work.Q{
			constants.SeriesIdKey: seriesId,
		})
		if err != nil {
			fmt.Println(err)
			return err
		} else {
			log.Println("Created job: ", updateJob.ID)
		}

	} else if transcodeType == constants.Movie {
		updateJob, err := Enqueuer.EnqueueUnique("update-radarr", work.Q{
			constants.MovieIdKey: id,
		})
		if err != nil {
			fmt.Println(err)
			return err
		} else {
			log.Println("Created job: ", updateJob.ID)
		}
	}
	return err
}

func WorkerPool() {
	log.Println("Starting worker pool")
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "my_app_namespace" is the Redis namespace
	// redisPool is a Redis pool
	pool := work.NewWorkerPool(Webhook{}, 20, constants.JobQueueNamespace, &storage.RedisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Webhook).Log)
	//pool.Middleware((*Context).FindCustomer)

	// Customize options:
	pool.JobWithOptions(constants.TranscodeJobType, work.JobOptions{
		Priority:       1,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 1,
	}, (*Webhook).DoTranscode)

	pool.JobWithOptions("update-sonarr", work.JobOptions{
		Priority:       2,
		MaxFails:       5,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, (*Webhook).UpdateTVShow)

	pool.JobWithOptions("update-radarr", work.JobOptions{
		Priority:       2,
		MaxFails:       5,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, (*Webhook).UpdateMovie)


	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
	log.Println("Worker pool stopped")
	os.Exit(0)
}
