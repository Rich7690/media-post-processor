package worker

import (
	"media-web/internal/constants"
	"media-web/internal/transcode"
	"media-web/internal/web"
	"path"

	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
)

// MovieScanner scans for movies to be converted into a consistent format
type MovieScanner interface {
	ScanForMovies() error
	SearchForMissingMovies() error
}

type movieScannerImpl struct {
	client    web.RadarrClient
	scheduler WorkScheduler
}

// NewMovieScanner creates a new instance of MovieScanner
func NewMovieScanner(client web.RadarrClient, scheduler WorkScheduler) MovieScanner {
	return movieScannerImpl{client: client, scheduler: scheduler}
}

func (m movieScannerImpl) SearchForMissingMovies() error {
	cmd, err := m.client.ScanForMissingMovies()
	if err == nil {
		log.Info().Int("id", cmd.ID).Msg("Scanned for missing movies")
	}
	return err
}

func (m movieScannerImpl) ScanForMovies() error {
	movies, err := m.client.GetAllMovies()

	if err != nil {
		return err
	}

	for i := 0; i < len(movies); i++ {
		movie := movies[i]
		if movie.Downloaded {
			should, reason, err := transcode.ShouldTranscode(transcode.VideoFileImpl{
				FilePath:        path.Join(movie.Path, movie.MovieFile.RelativePath),
				ContainerFormat: movie.MovieFile.MediaInfo.ContainerFormat,
				VideoCodec:      movie.MovieFile.MediaInfo.VideoFormat,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to determine if we should transcode")
			}
			if should {
				log.Debug().Str("reason", reason).Msg("Found movie in wrong format: " + movie.MovieFile.RelativePath)
				_, err := m.scheduler.EnqueueUnique(constants.TranscodeJobType, work.Q{
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
