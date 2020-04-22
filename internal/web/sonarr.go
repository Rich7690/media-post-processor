package web

import (
	"fmt"
	"media-web/internal/config"
	"media-web/internal/utils"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"
)

type SonarrClient interface {
	GetAllEpisodeFiles(seriesId int) ([]SonarrEpisodeFile, error)
	GetAllSeries() ([]Series, error)
	CheckSonarrCommand(id int) (*SonarrCommand, error)
	RescanSeries(id int64) (*SonarrCommand, error)
	LookupTVEpisode(id int64) (*SonarrEpisodeFile, error)
	GetEpisodeFilePath(id int64) (string, int, error)
}

type SonarrClientImpl struct {
	webClient          utils.WebClient
	BaseSonarrEndpoint url.URL
}

func GetSonarrClient() SonarrClient {
	var endpoint url.URL
	if config.GetConfig().SonarrBaseEndpoint != nil {
		endpoint = *config.GetConfig().SonarrBaseEndpoint
	}
	return SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: endpoint,
	}
}

func (c SonarrClientImpl) sonarrGetRequest(path string, query url.Values, respBody interface{}) error {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.GetRequest(c.BaseSonarrEndpoint, path, query, respBody)
}

func (c SonarrClientImpl) sonarrPostRequest(path string, query url.Values, body interface{}, respBody interface{}) error {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.PostRequest(c.BaseSonarrEndpoint, path, query, body, respBody)
}

func (c SonarrClientImpl) GetAllEpisodeFiles(seriesId int) ([]SonarrEpisodeFile, error) {
	response := make([]SonarrEpisodeFile, 0)
	vals := url.Values{}
	vals.Add("seriesId", strconv.Itoa(seriesId))
	err := c.sonarrGetRequest("api/episodeFile", vals, &response)

	if err == utils.NotFoundError {
		return response, nil
	}
	return response, err
}

func (c SonarrClientImpl) GetAllSeries() ([]Series, error) {
	response := make([]Series, 0)
	err := c.sonarrGetRequest("api/series/", url.Values{}, &response)

	if err == utils.NotFoundError {
		return response, nil
	}
	return response, err
}

func (c SonarrClientImpl) CheckSonarrCommand(id int) (*SonarrCommand, error) {
	var response SonarrCommand
	err := c.sonarrGetRequest(fmt.Sprintf("api/command/%d", id), url.Values{}, &response)
	return &response, err
}

func (c SonarrClientImpl) RescanSeries(id int64) (*SonarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanSeries"
	payload["seriesId"] = id

	var response SonarrCommand
	err := c.sonarrPostRequest("api/command", url.Values{}, payload, &response)

	return &response, err
}

func (c SonarrClientImpl) LookupTVEpisode(id int64) (*SonarrEpisodeFile, error) {
	var response SonarrEpisodeFile
	err := c.sonarrGetRequest(fmt.Sprintf("api/episodeFile/%d", id), url.Values{}, &response)
	if err == utils.NotFoundError {
		return nil, nil
	}
	return &response, err
}

func (c SonarrClientImpl) GetEpisodeFilePath(id int64) (string, int, error) {
	episodeFile, err := c.LookupTVEpisode(id)
	if err != nil {
		return "", -1, err
	}
	if episodeFile != nil {
		return episodeFile.Path, episodeFile.SeriesID, nil
	} else {
		log.Warn().Msg("Could not find episodeFile")
	}
	return "", -1, nil
}
