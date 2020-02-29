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
)

type RadarrClient interface {
	CheckRadarrCommand(id int) (*RadarrCommand, error)
	RescanMovie(id int64) (*RadarrCommand, error)
	LookupMovie(id int64) (*RadarrMovie, error)
	GetAllMovies() ([]RadarrMovie, error)
	GetMovieFilePath(id int64) (string, error)
}

type RadarrClientImpl struct {
	webClient utils.WebClient
	RadarrBaseEndpoint url.URL
}

func GetRadarrClient() RadarrClient {
	var endpoint url.URL
	if config.GetConfig().RadarrBaseEndpoint != nil {
		endpoint = *config.GetConfig().RadarrBaseEndpoint
	}
	return RadarrClientImpl{
		webClient: utils.GetWebClient(),
		RadarrBaseEndpoint: endpoint,
	}
}

func (c RadarrClientImpl) radarrGetRequest(client utils.WebClient, path string, query url.Values) (*http.Response, []byte, error) {
	query.Add("apikey", config.GetConfig().RadarrApiKey)
	return client.MakeGetRequest(c.RadarrBaseEndpoint, path, query)
}

func (c RadarrClientImpl) radarrPostRequest(client utils.WebClient, path string, query url.Values, body interface{}) (*http.Response, []byte, error) {
	query.Add("apikey", config.GetConfig().RadarrApiKey)
	resp, repBody, err := client.MakePostRequest(c.RadarrBaseEndpoint, path, query, body)
	if resp != nil && resp.StatusCode >= 300 {
		return resp, repBody, errors.New("got non-200 status code")
	}
	return resp, repBody, err
}

func (c RadarrClientImpl) CheckRadarrCommand(id int) (*RadarrCommand, error) {
	resp, body, err := c.radarrGetRequest(c.webClient, fmt.Sprintf("/api/command/%d", id), url.Values{})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 300 {
		response := RadarrCommand{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("failed to check command")
	}
}

func (c RadarrClientImpl) RescanMovie(id int64) (*RadarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanMovie"
	payload["movieId"] = id

	resp, value, err := c.radarrPostRequest(utils.WebClientImpl{}, "/api/command/", url.Values{}, payload)

	if err != nil {
		log.Error().Err(err).Msg("Error rescanning movie")
		return nil, err
	}

	if resp.StatusCode < 300 {
		response := RadarrCommand{}

		err = json.Unmarshal(value, &response)

		if err != nil {
			return nil, err
		}

		return &response, nil
	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(value)).Msg("Error calling radarr")
		return nil, errors.New("Bad response code")
	}

}

func (c RadarrClientImpl) LookupMovie(id int64) (*RadarrMovie, error) {

	resp, body, err := c.radarrGetRequest(utils.WebClientImpl{}, fmt.Sprintf("/api/movie/%d", id), url.Values{})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode < 300 {
		response := RadarrMovie{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Err(err).Str("response", string(body)).Msg("Error parsing json")
			return nil, err
		}

		return &response, nil

	} else {
		log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Error calling radarr")
		return nil, errors.New("Failed find movie")
	}
}

func (c RadarrClientImpl) GetAllMovies() ([]RadarrMovie, error) {
	response := make([]RadarrMovie, 0)
	resp, body, err := c.radarrGetRequest(utils.WebClientImpl{}, "/api/movie/", url.Values{})

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
		return nil, errors.New("Failed to get movies")
	}
}

func (c RadarrClientImpl) GetMovieFilePath(id int64) (string, error) {

	movie, err := c.LookupMovie(id)
	if err != nil {
		return "", err
	}
	if movie != nil {
		return movie.Path + "/" + movie.MovieFile.RelativePath, nil
	} else {
		log.Warn().Msg("Could not find movie from remote service")
	}

	return "", nil
}
