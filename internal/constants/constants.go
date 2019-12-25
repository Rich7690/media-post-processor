package constants

import "os"

var IsLocal = os.Getenv("LOCAL") == "true"

const SeriesIdKey = "seriesId"
const MovieIdKey = "movieId"
const TranscodeJobType = "transcode-job"
const JobQueueNamespace = "media-web"
const EpisodeFileIdKey = "episodeFileId"
const TranscodeTypeKey = "transcodeType"

type TranscodeType string

const (
	TV    TranscodeType = "TV"
	Movie               = "Movie"
)
