package worker

import (
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/web"
	"path/filepath"
)

type TVScanner struct {
	GetAllSeries func() ([]web.Series, error)
	GetAllEpisodeFiles func(seriesId int) ([]web.SonarrEpisodeFile, error)
}

func ScanForTVShows(scanner TVScanner, scheduler WorkScheduler) error {

	series, err := scanner.GetAllSeries()

	if err != nil {
		return err
	}

	for i := 0; i < len(series); i++ {
		log.Info().Msg("Scanning: " + series[i].Title)
		episodeFiles, err := scanner.GetAllEpisodeFiles(series[i].ID)
		if err != nil {
			log.Error().Err(err).Msg("Got error for series: " + series[i].Title)
		}
		for j := 0; j < len(episodeFiles); j++ {
			file := episodeFiles[j]
			ext := filepath.Ext(file.Path)

			if ext != ".mp4" {
				log.Info().Msg("Found episode file in wrong format: " + file.Path)
				_, err := scheduler.EnqueueUnique(constants.TranscodeJobType, work.Q{
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

