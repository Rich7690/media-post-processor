package worker

import (
	"context"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/transcode"
	"media-web/internal/web"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func ScanForTVShows(ctx context.Context, sonarrClient web.SonarrClient, wk storage.TranscodeWorker) {
	series, err := sonarrClient.GetAllSeries()

	if err != nil {
		return
	}

	for i := 0; i < len(series); i++ {
		log.Info().Msg("Scanning: " + series[i].Title)
		episodeFiles, err := sonarrClient.GetAllEpisodeFiles(series[i].ID)
		if err != nil {
			log.Error().Err(err).Msg("Got error for series: " + series[i].Title)
		}
		for j := 0; j < len(episodeFiles); j++ {
			file := episodeFiles[j]
			ext := filepath.Ext(file.Path)

			if ext != ".mp4" {
				log.Info().Msg("Found episode file in wrong format: " + file.Path)
				err := wk.EnqueueJob(ctx, &storage.TranscodeJob{
					TranscodeType: constants.TV,
					VideoFileImpl: transcode.VideoFileImpl{
						FilePath: file.Path,
						ContainerFormat: ext,
						VideoCodec: file.MediaInfo.VideoCodec,
					},
					VideoID: int64(file.SeriesID),
				})
				if err != nil {
					log.Error().Err(err).Msg("Error enqueueing tv transcode")
				}
			}
		}
	}
}
