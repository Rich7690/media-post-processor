package worker

import (
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/web"
	"path/filepath"
)

type MovieScanner struct {
	GetAllMovies  func() ([]web.RadarrMovie, error)
}

func ScanForMovies(scanner MovieScanner, scheduler WorkScheduler) error {

	movies, err := scanner.GetAllMovies()

	if err != nil {
		return err
	}

	for i := 0; i < len(movies); i++ {
		movie := movies[i]
		if movie.Downloaded {
			ext := filepath.Ext(movie.MovieFile.RelativePath)

			if ext != ".mp4" {
				log.Debug().Msg("Found movie in wrong format: " + movie.MovieFile.RelativePath)
				_, err := scheduler.EnqueueUnique(constants.TranscodeJobType, work.Q{
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

