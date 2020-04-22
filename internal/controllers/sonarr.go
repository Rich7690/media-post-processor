package controllers

import (
	"encoding/json"
	"media-web/internal/constants"
	"media-web/internal/web"
	"media-web/internal/worker"
	"net/http"

	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
)

func GetSonarrWebhookHandler(scheduler worker.WorkScheduler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body web.SonarrWebhook
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			log.Err(err).Msg("Failed to bind json")
			http.Error(w, "invalid json input", http.StatusBadRequest)
			return
		}

		switch body.EventType {
		case "Download":
			log.Info().Msg("Got Download request")
			job, err := scheduler.EnqueueUnique(constants.TranscodeJobType, work.Q{
				constants.EpisodeFileIdKey: body.EpisodeFile.ID,
				constants.TranscodeTypeKey: constants.TV,
			})

			if err != nil {
				log.Error().Err(err).Msg("Failed to enqueue work")
				http.Error(w, "failed to enqueue work", http.StatusInternalServerError)
				return
			}

			log.Info().Msgf("Enqueued job: %s", job.ID)
			break
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&body)
	}
}
