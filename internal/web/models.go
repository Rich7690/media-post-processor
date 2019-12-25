package web

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"time"
)


type SonarrCommand struct {
	Name                string    `json:"name"`
	StartedOn           time.Time `json:"startedOn"`
	StateChangeTime     time.Time `json:"stateChangeTime"`
	SendUpdatesToClient bool      `json:"sendUpdatesToClient"`
	State               string    `json:"state"`
	ID                  int       `json:"id"`
}

type RadarrCommand struct {
	Name string `json:"name"`
	Body struct {
		SendUpdatesToClient bool   `json:"sendUpdatesToClient"`
		UpdateScheduledTask bool   `json:"updateScheduledTask"`
		CompletionMessage   string `json:"completionMessage"`
		Name                string `json:"name"`
		Trigger             string `json:"trigger"`
	} `json:"body"`
	Priority            string    `json:"priority"`
	Status              string    `json:"status"`
	Queued              time.Time `json:"queued"`
	Trigger             string    `json:"trigger"`
	State               string    `json:"state"`
	Manual              bool      `json:"manual"`
	StartedOn           time.Time `json:"startedOn"`
	SendUpdatesToClient bool      `json:"sendUpdatesToClient"`
	UpdateScheduledTask bool      `json:"updateScheduledTask"`
	ID                  int       `json:"id"`
}

type JobData struct {
	TranscodeType constants.TranscodeType `json:"transcodeType"`
	Id            int                     `json:"id"`
}

type RadarrWebhook struct {
	EventType string `json:"eventType"`
	Movie     struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		ReleaseDate string `json:"releaseDate"`
	} `json:"movie"`
	RemoteMovie struct {
		TmdbID int    `json:"tmdbId"`
		ImdbID string `json:"imdbId"`
		Title  string `json:"title"`
		Year   int    `json:"year"`
	} `json:"remoteMovie"`
	MovieFile struct {
		ID             int    `json:"id"`
		RelativePath   string `json:"relativePath"`
		Path           string `json:"path"`
		Quality        string `json:"quality"`
		QualityVersion int    `json:"qualityVersion"`
		ReleaseGroup   string `json:"releaseGroup"`
	} `json:"movieFile"`
	IsUpgrade bool `json:"isUpgrade"`
}

func (r *RadarrWebhook) GetWebhookData(transcodeType constants.TranscodeType, id int64) error {
	result, err := redis.Bytes(storage.RedisPool.Get().Do("GET", fmt.Sprintf("%s-webhook-%d", transcodeType, id)))

	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = json.Unmarshal(result, &r)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

type SonarrWebhook struct {
	EventType string `json:"eventType"`
	Series    struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Path   string `json:"path"`
		TvdbID int    `json:"tvdbId"`
	} `json:"series"`
	Episodes []struct {
		ID             int       `json:"id"`
		EpisodeNumber  int       `json:"episodeNumber"`
		SeasonNumber   int       `json:"seasonNumber"`
		Title          string    `json:"title"`
		AirDate        string    `json:"airDate"`
		AirDateUtc     time.Time `json:"airDateUtc"`
		Quality        string    `json:"quality"`
		QualityVersion int       `json:"qualityVersion"`
	} `json:"episodes"`
	EpisodeFile struct {
		ID             int    `json:"id"`
		RelativePath   string `json:"relativePath"`
		Path           string `json:"path"`
		Quality        string `json:"quality"`
		QualityVersion int    `json:"qualityVersion"`
	} `json:"episodeFile"`
	IsUpgrade bool `json:"isUpgrade"`
}

type RadarrMovie struct {
	Title             string `json:"title"`
	AlternativeTitles []struct {
		SourceType string `json:"sourceType"`
		MovieID    int    `json:"movieId"`
		Title      string `json:"title"`
		SourceID   int    `json:"sourceId"`
		Votes      int    `json:"votes"`
		VoteCount  int    `json:"voteCount"`
		Language   string `json:"language"`
		ID         int    `json:"id"`
	} `json:"alternativeTitles"`
	SecondaryYearSourceID int       `json:"secondaryYearSourceId"`
	SortTitle             string    `json:"sortTitle"`
	SizeOnDisk            int64     `json:"sizeOnDisk"`
	Status                string    `json:"status"`
	Overview              string    `json:"overview"`
	InCinemas             time.Time `json:"inCinemas"`
	PhysicalRelease       time.Time `json:"physicalRelease"`
	Images                []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
	Website             string        `json:"website"`
	Downloaded          bool          `json:"downloaded"`
	Year                int           `json:"year"`
	HasFile             bool          `json:"hasFile"`
	YouTubeTrailerID    string        `json:"youTubeTrailerId"`
	Studio              string        `json:"studio"`
	Path                string        `json:"path"`
	ProfileID           int           `json:"profileId"`
	PathState           string        `json:"pathState"`
	Monitored           bool          `json:"monitored"`
	MinimumAvailability string        `json:"minimumAvailability"`
	IsAvailable         bool          `json:"isAvailable"`
	FolderName          string        `json:"folderName"`
	Runtime             int           `json:"runtime"`
	LastInfoSync        time.Time     `json:"lastInfoSync"`
	CleanTitle          string        `json:"cleanTitle"`
	ImdbID              string        `json:"imdbId"`
	TmdbID              int           `json:"tmdbId"`
	TitleSlug           string        `json:"titleSlug"`
	Genres              []interface{} `json:"genres"`
	Tags                []interface{} `json:"tags"`
	Added               time.Time     `json:"added"`
	Ratings             struct {
		Votes int     `json:"votes"`
		Value float64 `json:"value"`
	} `json:"ratings"`
	MovieFile struct {
		MovieID      int       `json:"movieId"`
		RelativePath string    `json:"relativePath"`
		Size         int64     `json:"size"`
		DateAdded    time.Time `json:"dateAdded"`
		SceneName    string    `json:"sceneName"`
		ReleaseGroup string    `json:"releaseGroup"`
		Quality      struct {
			Quality struct {
				ID         int    `json:"id"`
				Name       string `json:"name"`
				Source     string `json:"source"`
				Resolution string `json:"resolution"`
				Modifier   string `json:"modifier"`
			} `json:"quality"`
			CustomFormats []interface{} `json:"customFormats"`
			Revision      struct {
				Version int `json:"version"`
				Real    int `json:"real"`
			} `json:"revision"`
		} `json:"quality"`
		Edition   string `json:"edition"`
		MediaInfo struct {
			ContainerFormat              string  `json:"containerFormat"`
			VideoFormat                  string  `json:"videoFormat"`
			VideoCodecID                 string  `json:"videoCodecID"`
			VideoProfile                 string  `json:"videoProfile"`
			VideoCodecLibrary            string  `json:"videoCodecLibrary"`
			VideoBitrate                 int     `json:"videoBitrate"`
			VideoBitDepth                int     `json:"videoBitDepth"`
			VideoMultiViewCount          int     `json:"videoMultiViewCount"`
			VideoColourPrimaries         string  `json:"videoColourPrimaries"`
			VideoTransferCharacteristics string  `json:"videoTransferCharacteristics"`
			Width                        int     `json:"width"`
			Height                       int     `json:"height"`
			AudioFormat                  string  `json:"audioFormat"`
			AudioCodecID                 string  `json:"audioCodecID"`
			AudioCodecLibrary            string  `json:"audioCodecLibrary"`
			AudioAdditionalFeatures      string  `json:"audioAdditionalFeatures"`
			AudioBitrate                 int     `json:"audioBitrate"`
			RunTime                      string  `json:"runTime"`
			AudioStreamCount             int     `json:"audioStreamCount"`
			AudioChannels                int     `json:"audioChannels"`
			AudioChannelPositions        string  `json:"audioChannelPositions"`
			AudioChannelPositionsText    string  `json:"audioChannelPositionsText"`
			AudioProfile                 string  `json:"audioProfile"`
			VideoFps                     float64 `json:"videoFps"`
			AudioLanguages               string  `json:"audioLanguages"`
			Subtitles                    string  `json:"subtitles"`
			ScanType                     string  `json:"scanType"`
			SchemaRevision               int     `json:"schemaRevision"`
		} `json:"mediaInfo"`
		ID int `json:"id"`
	} `json:"movieFile"`
	QualityProfileID int `json:"qualityProfileId"`
	ID               int `json:"id"`
}

type SonarrEpisodeFile struct {
	SeriesID     int       `json:"seriesId"`
	SeasonNumber int       `json:"seasonNumber"`
	RelativePath string    `json:"relativePath"`
	Path         string    `json:"path"`
	Size         int       `json:"size"`
	DateAdded    time.Time `json:"dateAdded"`
	Quality      struct {
		Quality struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Resolution int    `json:"resolution"`
		} `json:"quality"`
		Revision struct {
			Version int `json:"version"`
			Real    int `json:"real"`
		} `json:"revision"`
	} `json:"quality"`
	MediaInfo struct {
		AudioChannels float64 `json:"audioChannels"`
		AudioCodec    string  `json:"audioCodec"`
		VideoCodec    string  `json:"videoCodec"`
	} `json:"mediaInfo"`
	QualityCutoffNotMet bool `json:"qualityCutoffNotMet"`
	ID                  int  `json:"id"`
}
