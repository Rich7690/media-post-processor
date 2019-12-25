package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/web"
	"media-web/internal/worker"
)

func RadarrWebhookHandler(c *gin.Context) {
	body := web.RadarrWebhook{}
	err := c.ShouldBindJSON(&body)

	if err != nil {
		log.Warn().Err(err).Msg("Invalid input")
		c.JSON(400, gin.H{
			"message": "Invalid input",
		})
		return
	}

	// Convert structs to JSON.
	data, err := json.Marshal(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal json")
	}
	log.Info().Interface("data", data).Msg("Got payload")

	switch body.EventType {
	case "Test":
		log.Info().Msg("Got Test request")
		if constants.IsLocal {
			bytes, err := json.Marshal(body)

			_, err = storage.RedisPool.Get().Do("SET", fmt.Sprintf("movie-webhook-%d", body.Movie.ID), bytes)

			if err != nil {
				log.Err(err).Msg("Failed to set redis value")
				c.JSON(500, gin.H{"message": "Failed to save body"})
				return
			}

			job, err := worker.Enqueuer.EnqueueUnique("test_job", work.Q{"id": body.Movie.ID})

			if err != nil {
				log.Error().Err(err).Msg("Failed to enqueue work")
				c.JSON(500, gin.H{"message": "Failed to enqueue work"})
				return
			}

			log.Printf("Enqueued job: %s", job.ID)
		}
		break
	case "Download":
		log.Info().Msg("Got Download request")

		job, err := worker.Enqueuer.EnqueueUnique(constants.TranscodeJobType, work.Q{
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
