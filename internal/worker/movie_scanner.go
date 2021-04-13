package worker

import (
	"context"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/transcode"
	"media-web/internal/web"
	"path"

	"github.com/rs/zerolog/log"
)

// MovieScanner scans for movies to be converted into a consistent format
type MovieScanner interface {
	ScanForMovies(ctx context.Context) error
	SearchForMissingMovies() error
}

type movieScannerImpl struct {
	client    web.RadarrClient
	tWorker   storage.TranscodeWorker
}

// NewMovieScanner creates a new instance of MovieScanner
func NewMovieScanner(client web.RadarrClient, tWorker storage.TranscodeWorker) MovieScanner {
	return movieScannerImpl{client: client, tWorker: tWorker}
}

func (m movieScannerImpl) SearchForMissingMovies() error {
	cmd, err := m.client.ScanForMissingMovies()
	if err == nil {
		log.Info().Int("id", cmd.ID).Msg("Scanned for missing movies")
	}
	return err
}

func (m movieScannerImpl) ScanForMovies(ctx context.Context) error {
	movies, err := m.client.GetAllMovies()

	if err != nil {
		return err
	}

	for i := 0; i < len(movies); i++ {
		movie := movies[i]
		if movie.Downloaded {
			video := transcode.VideoFileImpl{
				FilePath:        path.Join(movie.Path, movie.MovieFile.RelativePath),
				ContainerFormat: movie.MovieFile.MediaInfo.ContainerFormat,
				VideoCodec:      movie.MovieFile.MediaInfo.VideoFormat,
			}
			should, reason, err := transcode.ShouldTranscode(video)
			if err != nil {
				log.Error().Err(err).Msg("Failed to determine if we should transcode")
				continue
			}
			if should {
				log.Debug().Str("reason", reason).Msg("Found movie in wrong format: " + movie.MovieFile.RelativePath)
				err := m.tWorker.EnqueueJob(ctx, &storage.TranscodeJob{
					TranscodeType: constants.Movie,
					VideoFileImpl: video,
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to enqueue movie transcode")
				}
			}
		}
	}

	return nil
}
