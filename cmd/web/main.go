package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/controllers"
	"media-web/internal/worker"
	"os"
	"os/signal"
)

type nullWriter struct {

}

func (w nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type writer struct {

}

func (w writer) Write(p []byte) (n int, err error) {
	log.Debug().Msg(string(p))
	return len(p), nil
}

type errorWriter struct {

}

func (w errorWriter) Write(p []byte) (n int, err error) {
	log.Error().Msg(string(p))
	return len(p), nil
}


func startWorker() {
	log.Info().Msg("Starting worker.")


	err := worker.WorkerPool()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start consumers")
	}
	log.Info().Msg("Consumers are running")

}

func startWebserver() {
	log.Info().Msg("Starting server.")
	gin.DefaultWriter = writer{}
	gin.DefaultErrorWriter = errorWriter{}
	r := gin.Default()

	r.GET("/health", controllers.HealthHandler)
	r.POST("/api/radarr/webhook", controllers.GetRadarrWebhookHandler(worker.Enqueuer))
	r.POST("/api/sonarr/webhook", controllers.SonarrWebhookHandler)

	err := r.Run()

	if err != nil {
		panic(err)
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	if constants.IsLocal {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	if config.EnableWeb() {
		go startWebserver()
	}

	if config.EnableWorker() {
		go startWorker()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan
	log.Debug().Msg("Exiting.")
	os.Exit(0)
}
