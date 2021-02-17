package main

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/transcode"
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

	ctx, cancel := context.WithCancel(context.Background())

	utils.RegisterMetrics()

	log.Info().Msg("Running")

	wk := storage.GetTranscodeWorker(nil)

	go func() {
		for {

			err := wk.HandleErrored(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				log.Err(err).Msg("Error scanning in progress jobs")
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	err := wk.EnqueueJob(ctx, &storage.TranscodeJob{
		TranscodeType: constants.Movie,
		VideoFileImpl: transcode.VideoFileImpl{
			FilePath:        "testpath",
			ContainerFormat: "contformat",
			VideoCodec:      "codec",
		},
	})

	if err != nil {
		log.Fatal().Err(err).Msg("failed to enqueuue")
	}

	for {
		err = wk.DequeueJob(ctx, func(ctx context.Context, job storage.TranscodeJob) error {
			log.Debug().Interface("job", job).Msg("Got job")

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
