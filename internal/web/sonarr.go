package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/utils"
	"net/http"
	"net/url"
	"strconv"
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
	webClient utils.WebClient
	BaseSonarrEndpoint url.URL
}

func GetSonarrClient() SonarrClient {
	var endpoint url.URL
	if config.GetConfig().SonarrBaseEndpoint != nil {
		endpoint = *config.GetConfig().SonarrBaseEndpoint
	}
	return SonarrClientImpl{
		webClient: utils.GetWebClient(),
		BaseSonarrEndpoint: endpoint,
	}
}


func (c SonarrClientImpl) sonarrGetRequest( path string, query url.Values) (*http.Response, []byte, error) {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.MakeGetRequest(c.BaseSonarrEndpoint, path, query)
}

func (c SonarrClientImpl) sonarrPostRequest(path string, query url.Values, body interface{}) (*http.Response, []byte, error) {
	query.Add("apikey", config.GetConfig().SonarrApiKey)
	return c.webClient.MakePostRequest(c.BaseSonarrEndpoint, path, query, body)
}

func (c SonarrClientImpl) GetAllEpisodeFiles(seriesId int) ([]SonarrEpisodeFile, error) {
	response := make([]SonarrEpisodeFile, 0)
	vals := url.Values{}
	vals.Add("seriesId", strconv.Itoa(seriesId))
	resp, body, err := c.sonarrGetRequest( "/api/episodeFile", vals)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return response, nil
	}

	if resp.StatusCode < 300 {

		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("Failed to get episode files")
	}
}

func (c SonarrClientImpl) GetAllSeries() ([]Series, error) {
	response := make([]Series, 0)
	resp, body, err := c.sonarrGetRequest( "/api/series/", url.Values{})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return response, nil
	}

	if resp.StatusCode < 300 {

		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("Failed to get series")
	}
}

func (c SonarrClientImpl) CheckSonarrCommand(id int) (*SonarrCommand, error) {
	resp, body, err := c.sonarrGetRequest(fmt.Sprintf("/api/command/%d", id), url.Values{})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 300 {
		response := SonarrCommand{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("Failed to check command")
	}
}

func (c SonarrClientImpl) RescanSeries(id int64) (*SonarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanSeries"
	payload["seriesId"] = id

	resp, body, err := c.sonarrPostRequest( "/api/command", url.Values{}, payload)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 300 {
		response := SonarrCommand{}

		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil
	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, nil
	}
}

func (c SonarrClientImpl) LookupTVEpisode(id int64) (*SonarrEpisodeFile, error) {
	resp, body, err := c.sonarrGetRequest(fmt.Sprintf("/api/episodeFile/%d", id), url.Values{})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode < 300 {
		response := SonarrEpisodeFile{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("Failed find episode file")
	}
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
