package worker

import (
	"fmt"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/transcode"
	"os"
	"path"

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

	var videoFile transcode.VideoFile
	var id int64
	var err error
	var seriesID int
	switch transcodeType {
	case constants.TV:
		id = job.ArgInt64(constants.EpisodeFileIDKey)
		videoFile, seriesID, err = c.SonarrClient.GetEpisodeFilePath(id)
	case constants.Movie:
		id = job.ArgInt64(constants.MovieIDKey)
		videoFile, err = c.RadarrClient.GetMovieFilePath(id)
	default:
		log.Warn().Msg("Unknown transcodeType: " + string(transcodeType))
		return nil
	}

	if err != nil {
		log.Error().Err(err).Msg("Error getting input file path")
		return err
	}

	should, reason, err := transcode.ShouldTranscode(videoFile)
	if err == transcode.ErrFileNotExists || !should {
		log.Info().Str("path", videoFile.GetFilePath()).Msg("Not transcoding file")
		return nil
	}
	if err != nil {
		return err
	}

	ext := path.Ext(videoFile.GetFilePath())
	fileName := path.Base(videoFile.GetFilePath())
	baseDir := path.Dir(videoFile.GetFilePath())
	newPath := path.Join(baseDir, strings.Replace(fileName, ext, ".mp4", 1))
	log.Info().Str("reason", reason).Str("path", videoFile.GetFilePath()).Str("newPath", newPath).Msg("Transcoding file")

	trans = trans.Input(videoFile.GetFilePath()).Output(newPath)

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
		return errors.Wrap(err, "failed to start transcode")
	}

	start := 0
	var prog float64 = 0
	// Example of printing transcoding progress
	for msg := range progress {
		message := "Transcoding: " + videoFile.GetFilePath() + " -> " + fmt.Sprint(msg)
		if int(msg.GetProgress()) >= (20 + start) {
			log.Debug().Float64("progress", msg.GetProgress()).Msg("Transcoding: " + videoFile.GetFilePath())
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
		err = os.Remove(videoFile.GetFilePath())
	}

	if err != nil {
		log.Error().Err(err).Msg("Error deleting old file")
	}

	log.Info().Msg("Done transcoding: " + newPath)

	if transcodeType == constants.TV {
		var updateJob, err = c.Enqueuer.EnqueueUnique("update-sonarr", work.Q{
			constants.SeriesIDKey: seriesID,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to enqueue update job")
		}
		log.Debug().Msg("Created job: " + updateJob.ID)
	} else if transcodeType == constants.Movie {
		var updateJob, err = c.Enqueuer.EnqueueUnique("update-radarr", work.Q{
			constants.MovieIDKey: id,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to enqueue update job")
		}
		log.Debug().Msg("Created job: " + updateJob.ID)
	}
	return nil
}
