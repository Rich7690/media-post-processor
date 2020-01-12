package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/controllers"
	"media-web/internal/storage"
	"media-web/internal/worker"
	"os"
	"os/signal"
	"strings"
	"time"
)

type nullWriter struct {
}

func (w nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type writer struct {
}

func (w writer) Write(p []byte) (n int, err error) {
	line := string(p)
	log.Debug().Msg(strings.TrimSuffix(line, "\n"))
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
	worker.WorkerPool()
}

func startSonarrScanner() {
	log.Info().Msg("Starting Sonarr scanner")
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, os.Kill)
	repeat := make(chan bool, 1)
	repeat <- true // queue up first one to kick it off on start
	for {
		go func() {
			<-time.After(12 * time.Hour)
			repeat <- true
		}()

		select {
		case <-exitChan:
			return
		case <-repeat:
			log.Info().Msg("Scanning for TV in wrong format")
			err := worker.ScanForTVShows(worker.TVScannerImpl, worker.Enqueuer)

			log.Info().Msg("Done scanning for TV shows")
			if err != nil {
				log.Err(err).Msg("Error scanning for TV shows")
			}
			break
		}
	}
}

func startRadarrScanner() {
	log.Info().Msg("Starting Radarr scanner")
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, os.Kill)
	repeat := make(chan bool, 1)
	repeat <- true // queue up first one to kick it off on start
	for {
		go func() {
			<-time.After(12 * time.Hour)
			repeat <- true
		}()

		select {
		case <-exitChan:
			return
		case <-repeat:
			log.Info().Msg("Scanning for movies in wrong format")
			err := worker.ScanForMovies(worker.MovieScannerImpl, worker.Enqueuer)
			log.Info().Msg("Done scanning for movies")
			if err != nil {
				log.Err(err).Msg("Error scanning for movies")
			}
			break
		}
	}
}

func startWebserver() {
	log.Info().Msg("Starting server.")
	gin.DefaultWriter = writer{}
	gin.DefaultErrorWriter = errorWriter{}
	r := gin.Default()

	r.Use(static.ServeRoot("/", "./public"))
	r.GET("/health", controllers.HealthHandler)
	r.POST("/api/radarr/webhook", controllers.GetRadarrWebhookHandler(worker.Enqueuer))
	r.POST("/api/sonarr/webhook", controllers.GetSonarrWebhookHandler(worker.Enqueuer))
	r.GET("/api/movie/search", controllers.GetMoveSearchHandler())
	r.GET("/api/config", controllers.GetConfigHandler)

	err := r.Run()

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start web server")
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Level(zerolog.DebugLevel)
	if config.EnablePrettyLog() {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	storage.MigrateDB()

	if config.EnableWeb() {
		go startWebserver()
	}

	if config.EnableWorker() {
		go startWorker()
	}

	if config.EnableRadarrScanner() {
		go startRadarrScanner()
	}

	if config.EnableSonarrScanner() {
		go startSonarrScanner()
	}

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan
	log.Debug().Msg("Exiting.")
	os.Exit(0)
}
