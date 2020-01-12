package constants

import "os"

var IsLocal = os.Getenv("LOCAL") == "true"
var IsTest = os.Getenv("IS_TEST") == "true"

const SeriesIdKey = "seriesId"
const MovieIdKey = "movieId"
const TranscodeJobType = "transcode-job"
const UpdateRadarrJobName = "update-radarr"
const UpdateSonarrJobName = "update-sonarr"
const EpisodeFileIdKey = "episodeFileId"
const TranscodeTypeKey = "transcodeType"

type TranscodeType string

const (
	TV    TranscodeType = "TV"
	Movie               = "Movie"
)

type JobStatus string

const (
	Created    JobStatus = "Created"
	InProgress           = "InProgress"
	Completed            = "Completed"
	Failed               = "Failed"
)
