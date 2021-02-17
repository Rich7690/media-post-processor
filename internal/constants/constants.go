package constants

import "os"

var IsLocal = os.Getenv("LOCAL") == "true"

const SeriesIDKey = "seriesId"
const MovieIDKey = "movieId"
const TranscodeJobType = "transcode-job"
const UpdateRadarrJobName = "update-radarr"
const UpdateSonarrJobName = "update-sonarr"
const EpisodeFileIDKey = "episodeFileId"
const TranscodeTypeKey = "transcodeType"

type TranscodeType string

const (
	TV    TranscodeType = "TV"
	Movie TranscodeType = "Movie"
)
