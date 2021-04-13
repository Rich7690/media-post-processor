package main

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/storage"
	"media-web/internal/utils"
	"media-web/internal/web"

	//"media-web/internal/web"
	"media-web/internal/worker"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.StampMilli
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if config.GetConfig().EnablePrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli})
	}

	ctx, cancel := context.WithCancel(context.Background())

	utils.RegisterMetrics()

	log.Info().Msg("Running")

	wk := storage.GetTranscodeWorker()

	//t := worker.GetWorkerContext()

	worker.NewMovieScanner(web.GetRadarrClient(), wk).ScanForMovies(ctx)

	//if err != nil {
	//		log.Fatal().Err(err).Msg("Failed to scan")
	//	}

	for {
		err := wk.DequeueJob(ctx, func(ctx context.Context, job storage.TranscodeJob) error {
			log.Debug().Msg("Dequeued job")
			return nil
		})

		if err != nil {
			log.Err(err).Msg("failed to dequeue")
			time.Sleep(5 * time.Second)
		}
		if ctx.Err() != nil {
			break
		}
	}

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	cancel()
	log.Debug().Msg("Exiting.")
}
