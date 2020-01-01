package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/web"
	"media-web/internal/worker"
)

func SonarrWebhookHandler(c *gin.Context) {
	body := web.SonarrWebhook{}
	err := c.ShouldBindJSON(&body)

	if err != nil {
		log.Err(err).Msg("Failed to bind json")
		c.JSON(400, gin.H{
			"message": "Invalid input",
		})
		return
	}

	// Convert structs to JSON.
	data, err := json.Marshal(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal json")
	} else {
		log.Info().RawJSON("data", data).Msg("Encoded json")
	}

	switch body.EventType {
	case "Download":
		log.Info().Msg("Got Download request")
		job, err := worker.Enqueuer.EnqueueUnique(constants.TranscodeJobType, work.Q{
			constants.EpisodeFileIdKey: body.EpisodeFile.ID,
			constants.TranscodeTypeKey: constants.TV,
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to enqueue work")
			c.JSON(500, gin.H{"message": "Failed to enqueue work"})
			return
		}

		log.Info().Msg("Enqueued job: " + job.ID)
		break
	}

	c.JSON(200, body)
}
