package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/web"
	"media-web/internal/worker"
)

func GetRadarrWebhookHandler(scheduler worker.WorkScheduler) (func(c *gin.Context)) {
	return func (c *gin.Context) {
		body := web.RadarrWebhook{}
		err := c.ShouldBindJSON(&body)

		if err != nil {
			log.Warn().Err(err).Msg("Invalid input")
			c.JSON(400, gin.H{
				"message": "Invalid input",
			})
			return
		}

		switch body.EventType {
		case "Test":
			log.Info().Msg("Got Test request")
			break
		case "Download":
			log.Info().Msg("Got Download request")

			job, err := scheduler.EnqueueUnique(constants.TranscodeJobType, work.Q{
				constants.TranscodeTypeKey: constants.Movie,
				constants.MovieIdKey:       body.Movie.ID,
			})

			if err != nil {
				log.Error().Err(err).Msg("Failed to enqueue work")
				c.JSON(500, gin.H{"message": "Failed to enqueue work"})
				return
			}

			log.Printf("Enqueued job: %s", job.ID)
			break
		case "Rename":
			log.Info().Msg("Got Rename Request")
			break
		}

		c.JSON(200, body)
	}
}
