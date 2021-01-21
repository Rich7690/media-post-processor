package worker

import (
	"fmt"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"

	"github.com/floostack/transcoder"
	"github.com/floostack/transcoder/ffmpeg"
)

func GetTranscoder() transcoder.Transcoder {
	ffmpegConf := &ffmpeg.Config{
		FfmpegBinPath:   config.GetConfig().FfmpegPath,
		FfprobeBinPath:  config.GetConfig().FfprobePath,
		ProgressEnabled: true,
	}
	return ffmpeg.New(ffmpegConf)
}

func (c *WorkerContext) TranscodeTVShow() {

}

func (c *WorkerContext) TranscodeJobHandler(job *work.Job) error {
	// Create new instance of GetTranscoder
	trans := c.GetTranscoder()
	transcodeType := constants.TranscodeType(job.ArgString(constants.TranscodeTypeKey))

	var inputFilePath string
	var id int64
	var err error
	var seriesId int
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

	fileName := filepath.Base(inputFilePath)
	baseDir := filepath.Dir(inputFilePath)
	newPath := baseDir + "/" + strings.Replace(fileName, ext, ".mp4", 1)
	log.Debug().Msg("Transcoding to path: " + newPath)

	trans = trans.Input(inputFilePath).Output(newPath)

	log.Info().Msg("Transcoding: " + inputFilePath)

	preset := "veryfast"
	format := "mp4"
	codec := "libx264"
	tune := "film"
	var crf uint32 = 23
	opts := ffmpeg.Options{
		Preset:       &preset,
		OutputFormat: &format,
		VideoCodec:   &codec,
		Tune:         &tune,
		Crf:          &crf,
	}

	// Start transcoder process with progress checking
	progress, err := trans.Start(opts)

	// Returns a channel to get the transcoding progress
	if err != nil {
		return err
	}

	//now := time.Now()
	start := 0
	var prog float64 = 0
	// Example of printing transcoding progress
	for msg := range progress {
		message := "Transcoding: " + inputFilePath + " -> " + fmt.Sprint(msg)
		if int(msg.GetProgress()) >= (20 + start) {
			log.Debug().Float64("progress", msg.GetProgress()).Msg("Transcoding: " + inputFilePath)
			start = int(msg.GetProgress())
		}
		job.Checkin(message)
		prog = msg.GetProgress()
	}

	if prog != 100 {
		return errors.New("failed to get 100% progress on conversion. Keeping old file")
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
