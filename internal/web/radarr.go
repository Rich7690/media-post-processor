package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"media-web/internal/config"
	"media-web/internal/utils"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

type RadarrClient interface {
	CheckRadarrCommand(id int) (*RadarrCommand, error)
	RescanMovie(id int64) (*RadarrCommand, error)
	LookupMovie(id int64) (*RadarrMovie, error)
	GetAllMovies() ([]RadarrMovie, error)
	GetMovieFilePath(id int64) (string, error)
	ScanForMissingMovies() (*RadarrCommand, error)
}

type RadarrClientImpl struct {
	webClient          utils.WebClient
	RadarrBaseEndpoint url.URL
}

func GetRadarrClient() RadarrClient {
	var endpoint url.URL
	if config.GetConfig().RadarrBaseEndpoint != nil {
		endpoint = *config.GetConfig().RadarrBaseEndpoint
	}
	return RadarrClientImpl{
		webClient:          utils.GetWebClient(),
		RadarrBaseEndpoint: endpoint,
	}
}

func (c RadarrClientImpl) radarrGetRequest(client utils.WebClient, path string, query url.Values, respObject interface{}) error {
	query.Add("apikey", config.GetConfig().RadarrApiKey)
	return client.GetRequest(c.RadarrBaseEndpoint, path, query, respObject)
}

func (c RadarrClientImpl) radarrPostRequest(client utils.WebClient, path string, query url.Values, body interface{}) (*http.Response, []byte, error) {
	query.Add("apikey", config.GetConfig().RadarrApiKey)
	resp, repBody, err := client.MakePostRequest(c.RadarrBaseEndpoint, path, query, body)
	if resp != nil && resp.StatusCode >= 300 {
		return resp, repBody, errors.New("got non-200 status code")
	}
	return resp, repBody, err
}

func (c RadarrClientImpl) ScanForMissingMovies() (*RadarrCommand, error) {
	payload := make(map[string]interface{})

	payload["name"] = "missingMoviesSearch"
	payload["filterKey"] = "monitored"
	payload["filterValue"] = "true"

	resp, value, err := c.radarrPostRequest(utils.WebClientImpl{}, "api/command/", url.Values{}, payload)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		response := RadarrCommand{}

		err = json.Unmarshal(value, &response)

		if err != nil {
			return nil, err
		}

		return &response, nil
	}
	log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(value)).Msg("Error calling radarr")
	return nil, errors.New("bad response code")
}

func (c RadarrClientImpl) CheckRadarrCommand(id int) (*RadarrCommand, error) {
	var response RadarrCommand
	err := c.radarrGetRequest(c.webClient, fmt.Sprintf("api/command/%d", id), url.Values{}, &response)
	return &response, err
}

func (c RadarrClientImpl) RescanMovie(id int64) (*RadarrCommand, error) {
	payload := make(map[string]interface{})

	payload["name"] = "RescanMovie"
	payload["movieId"] = id

	resp, value, err := c.radarrPostRequest(utils.WebClientImpl{}, "api/command/", url.Values{}, payload)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 300 {
		response := RadarrCommand{}

		err = json.Unmarshal(value, &response)

		if err != nil {
			return nil, err
		}

		return &response, nil
	}
	log.Err(err).Int("status_code", resp.StatusCode).Str("response", string(value)).Msg("Error calling radarr")
	return nil, errors.New("Bad response code")
}

func (c RadarrClientImpl) LookupMovie(id int64) (*RadarrMovie, error) {
	var response RadarrMovie
	err := c.radarrGetRequest(utils.WebClientImpl{}, fmt.Sprintf("api/movie/%d", id), url.Values{}, &response)

	if err == utils.ErrNotFound {
		return nil, nil
	}
	return &response, err
}

func (c RadarrClientImpl) GetAllMovies() ([]RadarrMovie, error) {
	response := make([]RadarrMovie, 0)
	err := c.radarrGetRequest(utils.WebClientImpl{}, "api/movie/", url.Values{}, &response)

	if err == utils.ErrNotFound {
		return response, nil
	}
	return response, err
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
