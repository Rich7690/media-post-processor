package main

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if config.GetConfig().EnablePrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli})
	}

	_, cancel := context.WithCancel(context.Background())

	utils.RegisterMetrics()

	log.Info().Msg("Running")

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	cancel()
	log.Debug().Msg("Exiting.")
}
