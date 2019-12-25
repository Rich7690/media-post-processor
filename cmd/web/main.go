package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"media-web/internal/controllers"
	"media-web/internal/worker"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	if constants.IsLocal {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Info().Msg("Starting server")

	go worker.WorkerPool()

	r := gin.Default()

	r.GET("/health", controllers.HealthHandler)
	r.POST("/api/radarr/webhook", controllers.RadarrWebhookHandler)
	r.POST("/api/sonarr/webhook", controllers.SonarrWebhookHandler)

	err := r.Run()

	if err != nil {
		panic(err)
	}
}
