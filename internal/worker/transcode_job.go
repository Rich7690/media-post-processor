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
	"media-web/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

func (c *WorkerContext) TranscodeJobHandler(job *work.Job) error {
	// Create new instance of GetTranscoder
	trans := c.GetTranscoder()
	transcodeType := constants.TranscodeType(job.ArgString(constants.TranscodeTypeKey))

	var inputFilePath = ""
	var id int64 = -1
	var err error = nil
	var seriesId = -1
	switch transcodeType {
	case constants.TV:
		id = job.ArgInt64(constants.EpisodeFileIdKey)
		inputFilePath, seriesId, err = c.SonarrClient.GetEpisodeFilePath(id)
	case constants.Movie:
		id = job.ArgInt64(constants.MovieIdKey)
		inputFilePath, err = c.RadarrClient.GetMovieFilePath(id)
	default:
		log.Warn().Msg("Unknown transcodeType: " + string(transcodeType))
		return nil
	}

	if err != nil {
		log.Error().Err(err).Msg("Error getting input file path")
		return err
	}
	if constants.IsLocal {
		inputFilePath = "/Users/unknowndev/Downloads/test.mkv"
	}

	if inputFilePath != "" {
		log.Info().Msg("Working on transcode at path: " + inputFilePath)
	} else {
		log.Warn().Msg("Could not get input file path")
		return nil
	}

	if !utils.FileExists(inputFilePath) {
		log.Warn().Msg("Could not find file at path: " + inputFilePath)
		return nil
	}
	// Initialize GetTranscoder passing the input file path and output file path
	ext := filepath.Ext(inputFilePath)

	if ext == ".mp4" {
		log.Debug().Msg("File is already mp4 extension. Skipping...")
		return nil
	} else {
		log.Debug().Msg("Current extension: " + ext)
	}

	trans.SetConfiguration(ffmpeg.Configuration{
		FfmpegBin:  config.GetConfig().FfmpegPath,
		FfprobeBin: config.GetConfig().FfprobePath,
	})

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

	// Returns a channel to get the transcoding progress
	progress := trans.Output()

	now := time.Now()
	// Example of printing transcoding progress
	for msg := range progress {
		message := "Transcoding: " + inputFilePath + " -> " + fmt.Sprint(msg)
		if time.Since(now) > (30 * time.Second) {
			log.Debug().Float64("progress", msg.Progress).Msg("Transcoding: " + inputFilePath)
			now = time.Now()
		}
		job.Checkin(message)
	}

	// This channel is used to wait for the transcoding process to end
	err = <-done

	if err != nil {
		log.Error().Err(err).Msg("Error performing transcode")
		return err
	}

	log.Info().Msg("Deleting old file")

	if !constants.IsLocal {
		err = os.Remove(inputFilePath)
	}

	if err != nil {
		log.Error().Err(err).Msg("Error deleting old file")
	}

	log.Info().Msg("Done transcoding: " + newPath)

	if transcodeType == constants.TV {
		updateJob, err := c.Enqueuer.EnqueueUnique("update-sonarr", work.Q{
			constants.SeriesIdKey: seriesId,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to enqueue update job")
		} else {
			log.Debug().Msg("Created job: " + updateJob.ID)
		}

	} else if transcodeType == constants.Movie {
		updateJob, err := c.Enqueuer.EnqueueUnique("update-radarr", work.Q{
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

