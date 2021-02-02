package web

import (
	"fmt"
	"media-web/internal/config"
	"media-web/internal/transcode"
	"media-web/internal/utils"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"
)

type SonarrClient interface {
	GetAllEpisodeFiles(seriesID int) ([]SonarrEpisodeFile, error)
	GetAllSeries() ([]Series, error)
	CheckSonarrCommand(id int) (*SonarrCommand, error)
	RescanSeries(id *int64) (*SonarrCommand, error)
	LookupTVEpisode(id int64) (*SonarrEpisodeFile, error)
	GetEpisodeFilePath(id int64) (transcode.VideoFile, int, error)
}

type SonarrClientImpl struct {
	webClient utils.WebClient
}

func GetSonarrClient() SonarrClient {
	var endpoint *url.URL
	if config.GetConfig().SonarrBaseEndpoint != nil {
		endpoint = config.GetConfig().SonarrBaseEndpoint
	} else {
		log.Fatal().Msg("No base sonarr endpoint specified")
	}
	return &SonarrClientImpl{
		webClient: utils.GetWebClient(endpoint),
	}
}

func (c *SonarrClientImpl) sonarrGetRequest(path string, query url.Values, respBody interface{}) error {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.GetRequest(path, query, respBody)
}

func (c *SonarrClientImpl) sonarrPostRequest(path string, query url.Values, body, respBody interface{}) error {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.PostRequest(path, query, body, respBody)
}

func (c *SonarrClientImpl) GetAllEpisodeFiles(seriesID int) ([]SonarrEpisodeFile, error) {
	response := make([]SonarrEpisodeFile, 0)
	vals := url.Values{}
	vals.Add("seriesId", strconv.Itoa(seriesID))
	err := c.sonarrGetRequest("api/episodeFile", vals, &response)

	if err == utils.ErrNotFound {
		return response, nil
	}
	return response, err
}

func (c *SonarrClientImpl) GetAllSeries() ([]Series, error) {
	response := make([]Series, 0)
	err := c.sonarrGetRequest("api/series/", url.Values{}, &response)

	if err == utils.ErrNotFound {
		return response, nil
	}
	return response, err
}

func (c *SonarrClientImpl) CheckSonarrCommand(id int) (*SonarrCommand, error) {
	var response SonarrCommand
	err := c.sonarrGetRequest(fmt.Sprintf("api/command/%d", id), url.Values{}, &response)
	return &response, err
}

func (c *SonarrClientImpl) RescanSeries(id *int64) (*SonarrCommand, error) {
	payload := make(map[string]interface{})

	payload["name"] = "RescanSeries"
	if id != nil {
		payload["seriesId"] = id
	}

	var response SonarrCommand
	err := c.sonarrPostRequest("api/command", url.Values{}, payload, &response)

	return &response, err
}

func (c *SonarrClientImpl) LookupTVEpisode(id int64) (*SonarrEpisodeFile, error) {
	var response SonarrEpisodeFile
	err := c.sonarrGetRequest(fmt.Sprintf("api/episodeFile/%d", id), url.Values{}, &response)
	if err == utils.ErrNotFound {
		return nil, nil
	}
	return &response, err
}

func (c *SonarrClientImpl) GetEpisodeFilePath(id int64) (transcode.VideoFile, int, error) {
	file := transcode.VideoFileImpl{}
	episodeFile, err := c.LookupTVEpisode(id)
	if err != nil {
		return file, -1, err
	}
	if episodeFile != nil {
		file.FilePath = episodeFile.Path
		file.VideoCodec = episodeFile.MediaInfo.VideoCodec
		return file, episodeFile.SeriesID, nil
	}
	log.Warn().Msg("Could not find episodeFile")

	return file, -1, utils.ErrNotFound
}
