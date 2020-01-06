package worker

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/work"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/models"
	"github.com/xfrr/goffmpeg/transcoder"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/utils"
	"media-web/internal/web"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Webhook struct {
	EventType string `json:"eventType"`
}

type WorkScheduler struct {
	EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
}

var jobsEnqueued = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "jobs_enqueued",
		Help: "Number of Jobs enqueued",
	},
)

var jobsAttempted = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "jobs_attempted",
		Help: "Number of Jobs attempted",
	},
)

var worker = work.NewEnqueuer(constants.JobQueueNamespace, &storage.RedisPool)

var Enqueuer = WorkScheduler{
	EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		jobsEnqueued.Inc()
		return worker.EnqueueUnique(jobName, args)
	},
}

func (c *Webhook) CountJobsPerformed(job *work.Job, next work.NextMiddlewareFunc) error {
	jobsAttempted.Inc()
	return next()
}

func (c *Webhook) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Info().Str("jobId", job.ID).Msg("Starting job: " + job.ID)
	return next()
}

func getMovieFilePath(id int64) (string, error) {

	movie, err := web.LookupMovie(id)
	if err != nil {
		return "", err
	}
	if movie != nil {
		return movie.Path + "/" + movie.MovieFile.RelativePath, nil
	} else {
		log.Warn().Msg("Could not find movie from remote service")
	}

	return "", nil
}

func getEpisodeFilePath(id int64) (string, int, error) {
	episodeFile, err := web.LookupTVEpisode(id)
	if err != nil {
		return "", -1, err
	}
	if episodeFile != nil {
		return episodeFile.Path, episodeFile.SeriesID, nil
	} else {
		log.Warn().Msg("Could not find episodeFile")
	}
	return "", -1, nil
}

func ScanForMovies() error {

	movies, err := web.GetAllMovies()

	if err != nil {
		return err
	}

	for i := 0; i < len(movies); i++ {
		movie := movies[i]
		if movie.Downloaded {
			ext := filepath.Ext(movie.MovieFile.RelativePath)

			if ext != ".mp4" {
				log.Debug().Msg("Found movie in wrong format: " + movie.MovieFile.RelativePath)
				_, err := Enqueuer.EnqueueUnique(constants.TranscodeJobType, work.Q{
					constants.TranscodeTypeKey: constants.Movie,
					constants.MovieIdKey:       movie.ID,
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to enqueue movie transcode")
				}
			}
		}
	}

	return nil
}

func ScanForTVShows() error {

	series, err := web.GetAllSeries()

	if err != nil {
		return err
	}

	for i := 0; i < len(series); i++ {
		log.Info().Msg("Scanning: " + series[i].Title)
		episodeFiles, err := web.GetAllEpisodeFiles(series[i].ID)
		if err != nil {
			log.Error().Err(err).Msg("Got error for series: " + series[i].Title)
		}
		for j := 0; j < len(episodeFiles); j++ {
			file := episodeFiles[j]
			ext := filepath.Ext(file.Path)

			if ext != ".mp4" {
				log.Info().Msg("Found episode file in wrong format: " + file.Path)
				_, err := Enqueuer.EnqueueUnique(constants.TranscodeJobType, work.Q{
					constants.TranscodeTypeKey: constants.TV,
					constants.EpisodeFileIdKey: file.ID,
				})
				if err != nil {
					log.Error().Err(err).Msg("Error enqueueing tv transcode")
				}
			}
		}
	}

	return nil
}

func (c *Webhook) UpdateTVShow(job *work.Job) error {

	seriesId := job.ArgInt64(constants.SeriesIdKey)

	cmd, err := web.RescanSeries(seriesId)

	if err != nil {
		log.Err(err).Msg("Error rescanning series")
		return err
	}

	for count := 0; count < 5; count++ {
		result, err := web.CheckSonarrCommand(cmd.ID)

		if err == nil {
			if strings.Index(result.State, "complete") != -1 {
				log.Info().Msg("Rescan complete for: " + strconv.Itoa(cmd.ID))
				return nil
			} else {
				log.Info().Msg("Rescan not complete yet for: " + strconv.Itoa(cmd.ID))
			}
		} else {
			log.Err(err).Msg("Error checking state of command: " + strconv.Itoa(cmd.ID))
		}

		time.Sleep(time.Second * 15)
	}

	return err
}

func (c *Webhook) UpdateMovie(job *work.Job) error {
	movieId := job.ArgInt64(constants.MovieIdKey)

	cmd, err := web.RescanMovie(movieId)

	if err != nil {
		log.Err(err).Msg("Error rescanning movie: " + strconv.Itoa(int(movieId)))
		return err
	}

	for count := 0; count < 5; count++ {
		result, err := web.CheckRadarrCommand(cmd.ID)

		if err == nil {
			if strings.Index(result.State, "complete") != -1 {
				log.Info().Msg("Rescan complete for: " + strconv.Itoa(cmd.ID))
				return nil
			} else {
				log.Info().Msg("Rescan not complete yet for: " + strconv.Itoa(cmd.ID))
			}
		} else {
			log.Err(err).Msg("Error checking status of command")
		}

		time.Sleep(time.Second * 15)
	}

	return nil
}

type Transcoder struct {
	SetConfiguration func(config ffmpeg.Configuration)
	Initialize       func(input string, output string) error
	MediaFile        func() *models.Mediafile
	Output           func() <-chan models.Progress
	Run              func(progress bool) <-chan error
}

func GetTranscoder() Transcoder {
	trans := transcoder.Transcoder{}
	return Transcoder{
		SetConfiguration: func(config ffmpeg.Configuration) {
			trans.SetConfiguration(config)
		},
		Initialize: func(input string, output string) error {
			return trans.Initialize(input, output)
		},
		MediaFile: func() *models.Mediafile {
			return trans.MediaFile()
		},
		Output: func() <-chan models.Progress {
			return trans.Output()
		},
		Run: func(progress bool) <-chan error {
			return trans.Run(progress)
		},
	}
}

func doTranscode(trans Transcoder, job *work.Job) error {
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
		log.Warn().Msg("Unknown transcodeType: " + string(transcodeType))
		return nil
	}

	if err != nil {
		log.Error().Err(err).Msg("Error getting input file path")
		return err
	}

	if inputFilePath != "" {
		log.Info().Msg("Working on transcode at path: " + inputFilePath)
	} else {
		log.Warn().Msg("Could not get input file path")
		return nil
	}

	trans.SetConfiguration(ffmpeg.Configuration{
		FfmpegBin:  config.GetFfmpegPath(),
		FfprobeBin: config.GetFfprobePath(),
	})

	if !utils.FileExists(inputFilePath) {
		log.Warn().Msg("Could not find file at path: " + inputFilePath)
		return nil
	}
	// Initialize transcoder passing the input file path and output file path
	ext := filepath.Ext(inputFilePath)

	if ext == ".mp4" {
		log.Debug().Msg("File is already mp4 extension. Skipping...")
		return nil
	} else {
		log.Debug().Msg("Current extension: " + ext)
	}

	fileName := filepath.Base(inputFilePath)
	baseDir := filepath.Dir(inputFilePath)
	newPath := baseDir + "/" + strings.Replace(fileName, ext, ".mp4", 1)
	log.Debug().Msg("Transcoding to path: " + newPath)
	err = trans.Initialize(inputFilePath, newPath)

	if err != nil {
		log.Err(err).Msg("Error initializing transcode")
		return err
	}

	log.Info().Msg("Transcoding: " + trans.MediaFile().InputPath())

	trans.MediaFile().SetPreset("veryfast")
	trans.MediaFile().SetOutputFormat("mp4")
	trans.MediaFile().SetVideoCodec("libx264")
	trans.MediaFile().SetQuality(23)
	trans.MediaFile().SetTune("film")

	bytes, err := json.Marshal(trans.MediaFile())

	if err == nil {
		log.Debug().Msg("Json: " + string(bytes))
	} else {
		log.Error().Err(err).Msg("Failed to marshal media file")
	}

	log.Debug().Msg("Running ffmpeg command: \"" + strings.Join(trans.MediaFile().ToStrCommand(), " ") + "\"")

	// Start transcoder process with progress checking
	done := trans.Run(true)
	//done := make(chan error, 1)
	//done <- nil

	// Returns a channel to get the transcoding progress
	progress := trans.Output()

	// Example of printing transcoding progress
	for msg := range progress {
		message := "Transcoding: " + inputFilePath + " -> " + fmt.Sprint(msg)
		log.Debug().Float64("progress", msg.Progress).Msg("Transcoding: " + inputFilePath)
		job.Checkin(message)
	}

	// This channel is used to wait for the transcoding process to end
	err = <-done

	if err != nil {
		log.Error().Err(err).Msg("Error performing transcode")
		return err
	}

	log.Info().Msg("Deleting old file")

	err = os.Remove(inputFilePath)

	if err != nil {
		log.Error().Err(err).Msg("Error deleting old file")
	}

	log.Info().Msg("Done transcoding: " + newPath)

	if transcodeType == constants.TV {
		updateJob, err := Enqueuer.EnqueueUnique("update-sonarr", work.Q{
			constants.SeriesIdKey: seriesId,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to enqueue update job")
		} else {
			log.Debug().Msg("Created job: " + updateJob.ID)
		}

	} else if transcodeType == constants.Movie {
		updateJob, err := Enqueuer.EnqueueUnique("update-radarr", work.Q{
			constants.MovieIdKey: id,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to enqueue update job")
		} else {
			log.Debug().Msg("Created job: " + updateJob.ID)
		}
	}
	return err
}

func (c *Webhook) TranscodeJobHandler(job *work.Job) error {
	// Create new instance of transcoder
	trans := GetTranscoder()
	return doTranscode(trans, job)
}

func WorkerPool() {
	log.Info().Msg("Starting worker pool")
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "my_app_namespace" is the Redis namespace
	// redisPool is a Redis pool
	pool := work.NewWorkerPool(Webhook{}, 20, constants.JobQueueNamespace, &storage.RedisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Webhook).Log)
	pool.Middleware((*Webhook).CountJobsPerformed)
	//pool.Middleware((*Context).FindCustomer)

	// Customize options:
	pool.JobWithOptions(constants.TranscodeJobType, work.JobOptions{
		Priority:       1,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 1,
	}, (*Webhook).TranscodeJobHandler)

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
	log.Info().Msg("Worker pool stopped")
	os.Exit(0)
}
