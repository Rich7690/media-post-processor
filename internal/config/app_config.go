package config

import (
	"net/url"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	EnableWeb           bool     `env:"ENABLE_WEB" envDefault:"true"`
	EnableWorker        bool     `env:"ENABLE_WORKER" envDefault:"false"`
	EnableRadarrScanner bool     `env:"ENABLE_RADARR_SCANNER" envDefault:"false"`
	EnableSonarrScanner bool     `env:"ENABLE_SONARR_SCANNER" envDefault:"false"`
	EnablePrettyLog     bool     `env:"ENABLE_PRETTYLOG" envDefault:"false"`
	RadarrApiKey        string   `env:"RADARR_API_KEY"`
	SonarrApiKey        string   `env:"SONARR_API_KEY"`
	RadarrBaseEndpoint  *url.URL `env:"RADARR_BASE_ENDPOINT"`
	SonarrBaseEndpoint  *url.URL `env:"SONARR_BASE_ENDPOINT"`
	RedisAddress        *url.URL `env:"REDIS_ADDRESS"`
	JobQueueNamespace   string   `env:"JOB_QUEUE_NAMESPACE" envDefault:"media-web"`
	FfmpegPath          string   `env:"FFMPEG_PATH" envDefault:"/usr/bin/ffmpeg"`
	FfprobePath         string   `env:"FFPROBE_PATH" envDefault:"/usr/bin/ffprobe"`
	MovieScanCron       string   `env:"MOVIE_SCAN_CRON" envDefault:"0 0 * * *"`
	TVScanCron          string   `env:"TV_SCAN_CRON" envDefault:"0 1 * * *"`
}

var config = ValidateConfig()

func ValidateConfig() Config {
	cfg := Config{}
	funcs := make(map[reflect.Type]env.ParserFunc)
	funcs[reflect.TypeOf(&url.URL{})] = func(v string) (i interface{}, e error) {
		return url.Parse(v)
	}

	if err := env.ParseWithFuncs(&cfg, funcs); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}
	return cfg
}

func GetConfig() Config {
	return config
}
