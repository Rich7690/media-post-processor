package main

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/controllers"
	"media-web/internal/storage"
	"media-web/internal/utils"
	"media-web/internal/web"
	"media-web/internal/worker"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func startWorker(ctx context.Context) {
	log.Info().Msg("Starting worker.")

	wk := storage.GetTranscodeWorker()

	t := worker.GetWorkerContext()

	go func() {
		log.Info().Msg("Watching for jobs")
		for {
			err := wk.DequeueJob(ctx, t.HandleTranscodeJob)

			if err != nil {
				log.Err(err).Msg("failed to dequeue")
				time.Sleep(5 * time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Minute):
				wk.HandleErrored(ctx)
			}
		}
	}()

	worker.StartWorkerPool(ctx, worker.GetWorkerContext(), worker.WorkerPoolFactoryImpl{})
}

func performTVScan(ctx context.Context) {
	log.Info().Msg("Scanning for TV in wrong format")
	worker.ScanForTVShows(ctx, web.GetSonarrClient(), storage.GetTranscodeWorker())
	log.Info().Msg("Done scanning for TV shows")
}

func performScan(ctx context.Context, scanner worker.MovieScanner) {
	log.Info().Msg("Scanning for missing movies")
	err := scanner.SearchForMissingMovies()
	if err != nil {
		log.Err(err).Msg("Error searching for movies")
	}
	log.Info().Msg("Scanning for movies in wrong format")
	err = scanner.ScanForMovies(ctx)
	log.Info().Msg("Done scanning for movies")
	if err != nil {
		log.Err(err).Msg("Error scanning for movies")
	}
}

func startScanners(ctx context.Context) {
	c := cron.New()

	tWorker := storage.GetTranscodeWorker()
	if config.GetConfig().EnableRadarrScanner {
		scanner := worker.NewMovieScanner(web.GetRadarrClient(), tWorker)

		_, err := c.AddFunc(config.GetConfig().MovieScanCron, func() {
			performScan(ctx, scanner)
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start Radarr scanner")
		}
	}

	if config.GetConfig().EnableSonarrScanner {
		_, err := c.AddFunc(config.GetConfig().TVScanCron, func() {
			performTVScan(ctx)
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start Radarr scanner")
		}
	}
	c.Start()

	<-ctx.Done()
	c.Stop()
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("err", err).Str("path", req.URL.Path).Msg("Recovered from panic")
				http.Error(w, "Unknown error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

const Timeout = 4 * time.Second

func startWebserver(ctx context.Context) {
	log.Info().Msg("Starting server.")
	ro := mux.NewRouter()

	ro.StrictSlash(true)
	ro.HandleFunc("/health", controllers.HealthHandler)
	ro.HandleFunc("/api/radarr/webhook", controllers.GetRadarrWebhookHandler(worker.Enqueuer))
	ro.HandleFunc("/api/sonarr/webhook", controllers.GetSonarrWebhookHandler(worker.Enqueuer)).Methods(http.MethodPost)
	ro.Handle("/metrics", promhttp.Handler())
	ro.HandleFunc("/debug/pprof/", pprof.Index).Methods("GET")
	ro.HandleFunc("/debug/pprof/{name}", pprofHandler())

	serv := http.Server{
		Addr:         ":8080",
		Handler:      recoverHandler(http.TimeoutHandler(ro, Timeout, "Failed to handle request in time")),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		<-ctx.Done()
		err := serv.Close()
		if err != nil {
			log.Err(err).Msg("error on server close")
		}
	}()
	err := serv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start web server")
	}
}

func pprofHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if name == "cmdline" {
			pprof.Cmdline(w, r)
			return
		}
		pprof.Handler(name).ServeHTTP(w, r)
	}
}

func main() {
	zerolog.TimeFieldFormat = time.StampMilli
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if config.GetConfig().EnablePrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli})
	}

	ctx, cancel := context.WithCancel(context.Background())

	utils.RegisterMetrics()
	name, _ := os.Hostname()
	log.Info().Str("hostname", name).Msg("Starting web")

	go startWebserver(ctx)

	if config.GetConfig().EnableWorker {
		go startWorker(ctx)
	}

	go startScanners(ctx)

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	cancel()
	log.Debug().Msg("Exiting.")
}
