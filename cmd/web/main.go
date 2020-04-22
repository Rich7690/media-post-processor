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
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func startWorker(ctx context.Context) {
	log.Info().Msg("Starting worker.")
	worker.StartWorkerPool(worker.GetWorkerContext(), worker.WorkerPoolFactoryImpl{}, ctx)
}

func startSonarrScanner(ctx context.Context) {
	log.Info().Msg("Starting Sonarr scanner")
	time.Sleep(10 * time.Second)
	repeat := make(chan bool, 1)
	repeat <- true // queue up first one to kick it off on start
	for {
		go func() {
			<-time.After(12 * time.Hour)
			repeat <- true
		}()

		select {
		case <-repeat:
			log.Info().Msg("Scanning for TV in wrong format")
			worker.ScanForTVShows(web.GetSonarrClient(), worker.Enqueuer)
			log.Info().Msg("Done scanning for TV shows")
			break
		case <-ctx.Done():
			log.Info().Msg("Closing Sonarr scanner")
			return
		}
	}
}

func startRadarrScanner(ctx context.Context) {
	log.Info().Msg("Starting Radarr scanner")
	time.Sleep(10 * time.Second)
	scanner := worker.NewMovieScanner(web.GetRadarrClient(), worker.Enqueuer)
	repeat := make(chan bool, 1)
	repeat <- true // queue up first one to kick it off on start
	for {
		go func() {
			<-time.After(12 * time.Hour)
			repeat <- true
		}()

		select {
		case <-repeat:
			log.Info().Msg("Scanning for movies in wrong format")
			err := scanner.ScanForMovies()
			log.Info().Msg("Done scanning for movies")
			if err != nil {
				log.Err(err).Msg("Error scanning for movies")
			}
			break
		case <-ctx.Done():
			log.Info().Msg("Closing Radarr scanner")
			return
		}
	}
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

	//r.Use(static.ServeRoot("/", "./public"))
	ro.HandleFunc("/health", controllers.HealthHandler)
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
		<- ctx.Done()
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

	if config.GetConfig().EnableRadarrScanner {
		go startRadarrScanner(ctx)
	}

	if config.GetConfig().EnableSonarrScanner {
		go startSonarrScanner(ctx)
	}

	log.Debug().Msg("Waiting for exit signal")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan
	cancel()
	log.Debug().Msg("Exiting.")
}
