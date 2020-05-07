package main

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/controllers"
	"media-web/internal/web"
	"media-web/internal/worker"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/robfig/cron"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func startWorker(ctx context.Context) {
	log.Info().Msg("Starting worker.")
	worker.StartWorkerPool(worker.GetWorkerContext(), worker.WorkerPoolFactoryImpl{}, ctx)
}

func performTVScan() {
	log.Info().Msg("Scanning for TV in wrong format")
	worker.ScanForTVShows(web.GetSonarrClient(), worker.Enqueuer)
	log.Info().Msg("Done scanning for TV shows")
}

func performScan(scanner worker.MovieScanner) {
	log.Info().Msg("Scanning for missing movies")
	err := scanner.SearchForMissingMovies()
	if err != nil {
		log.Err(err).Msg("Error searching for movies")
	}
	log.Info().Msg("Scanning for movies in wrong format")
	err = scanner.ScanForMovies()
	log.Info().Msg("Done scanning for movies")
	if err != nil {
		log.Err(err).Msg("Error scanning for movies")
	}
}

func startScanners(ctx context.Context) {
	c := cron.New()

	if config.GetConfig().EnableRadarrScanner {
		scanner := worker.NewMovieScanner(web.GetRadarrClient(), worker.Enqueuer)

		err := c.AddFunc("0 0 * * *", func() {
			performScan(scanner)
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start Radarr scanner")
		}
	}

	if config.GetConfig().EnableSonarrScanner {
		err := c.AddFunc("0 1 * * *", func() {
			performTVScan()
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

func startWebserver(ctx context.Context) {
	log.Info().Msg("Starting server.")
	ro := mux.NewRouter()

	ro.StrictSlash(true)
	//r.Use(static.ServeRoot("/", "./public"))
	ro.HandleFunc("/health", controllers.HealthHandler)
	/*ro.HandleFunc("/api/movies", mvCtl.HandleCreate).Methods("POST")
	ro.HandleFunc("/api/movies", mvCtl.HandleList).Methods("GET")
	ro.HandleFunc("/api/movies/{id}", mvCtl.HandleDelete).Methods("DELETE")
	ro.HandleFunc("/api/search/movies", mvCtl.HandleSearch).Methods("GET")*/
	ro.HandleFunc("/api/radarr/webhook", controllers.GetRadarrWebhookHandler(worker.Enqueuer))
	ro.HandleFunc("/api/sonarr/webhook", controllers.GetSonarrWebhookHandler(worker.Enqueuer)).Methods(http.MethodPost)
	ro.Handle("/metrics", promhttp.Handler())
	ro.HandleFunc("/debug/pprof/", pprof.Index).Methods("GET")
	ro.HandleFunc("/debug/pprof/{name}", pprofHandler())
	//r.GET("/api/config", controllers.GetConfigHandler)

	serv := http.Server{
		Addr:         ":8080",
		Handler:      recoverHandler(http.TimeoutHandler(ro, 4*time.Second, "Failed to handle request in time")),
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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.With().Timestamp().Logger()

	if config.GetConfig().EnablePrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli})
	}

	ctx, cancel := context.WithCancel(context.Background())

	go startWebserver(ctx)

	if config.GetConfig().EnableWorker {
		go startWorker(ctx)
	}

	go startScanners(ctx)

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan
	cancel()
	log.Debug().Msg("Exiting.")
}
