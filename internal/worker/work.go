package worker

import (
	"fmt"
	"github.com/gocraft/work"
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
	"strings"
)

type Webhook struct {
	EventType string `json:"eventType"`
}

type WorkScheduler struct {
	EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
}

var worker = work.NewEnqueuer(config.JobQueueNamespace(), &storage.RedisPool)

var Enqueuer = WorkScheduler{
	EnqueueUnique: worker.EnqueueUnique,
}

func (c *Webhook) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Info().Str("jobId", job.ID).Msg("Starting job: " + job.ID)
	return next()
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

func DoTranscode(trans Transcoder, job *work.Job) error {
	transcodeType := constants.TranscodeType(job.ArgString(constants.TranscodeTypeKey))

	var inputFilePath = ""
	var id int64 = -1
	var err error = nil
	var seriesId = -1
	switch transcodeType {
	case constants.TV:
		id = job.ArgInt64(constants.EpisodeFileIdKey)
		inputFilePath, seriesId, err = web.GetEpisodeFilePath(id)
	case constants.Movie:
		id = job.ArgInt64(constants.MovieIdKey)
		inputFilePath, err = web.GetMovieFilePath(id)
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
	return DoTranscode(trans, job)
}

func WorkerPool() {
	log.Info().Msg("Starting worker pool")
	pool := work.NewWorkerPool(Webhook{}, 20, config.JobQueueNamespace(), &storage.RedisPool)
	pool.Middleware((*Webhook).Log)

	pool.JobWithOptions(constants.TranscodeJobType, work.JobOptions{
		Priority:       1,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 1,
	}, (*Webhook).TranscodeJobHandler)

	pool.JobWithOptions(constants.UpdateSonarrJobName, work.JobOptions{
		Priority:       2,
		MaxFails:       5,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, (*Webhook).UpdateTVShow)

	pool.JobWithOptions(constants.UpdateRadarrJobName, work.JobOptions{
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
}
