package config

import "os"

var radarAPIKey = os.Getenv("RADARR_API_KEY")
var radarBaseEndpoint = os.Getenv("RADARR_BASE_ENDPOINT")
var sonarrAPIKey = os.Getenv("SONARR_API_KEY")
var sonarrBaseEndpoint = os.Getenv("SONARR_BASE_ENDPOINT")
var redisAddress = os.Getenv("REDIS_ADDRESS")
var ffmpegBin = os.Getenv("FFMPEG_PATH")
var ffprobeBin = os.Getenv("FFPROBE_PATH")

func GetFfmpegPath() string {
	if ffmpegBin != "" {
		return ffmpegBin
	}
	return "/usr/bin/ffmpeg"
}

func GetFfprobePath() string {
	if ffprobeBin != "" {
		return ffprobeBin
	}
	return "/usr/bin/ffprobe"
}

func GetRedisAddress() string {
	return redisAddress
}

func GetRadarAPIKey() string {
	return radarAPIKey
}

func GetRadarBaseEndpoint() string {
	return radarBaseEndpoint
}

func GetSonarrAPIKey() string {
	return sonarrAPIKey
}

func GetSonarrBaseEndpoint() string {
	return sonarrBaseEndpoint
}
