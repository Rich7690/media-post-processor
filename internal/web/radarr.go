package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/utils"
)

func CheckRadarrCommand(id int) (*RadarrCommand, error) {

	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/command/%d?apikey=%s", config.GetRadarBaseEndpoint(), id, config.GetRadarAPIKey()))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		response := RadarrCommand{}
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

func RescanMovie(id int64) (*RadarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanMovie"
	payload["movieId"] = id

	resp, value, err := utils.PostRequest(fmt.Sprintf("%s/api/command/?apikey=%s", config.GetRadarBaseEndpoint(), config.GetRadarAPIKey()), payload)

	if err != nil {
		log.Error().Err(err).Msg("Error rescanning movie")
		return nil, err
	}

	if resp.StatusCode == 200 {
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

func LookupMovie(id int64) (*RadarrMovie, error) {

	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/movie/%d?apikey=%s", config.GetRadarBaseEndpoint(), id, config.GetRadarAPIKey()))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode == 200 {
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

func GetAllMovies() ([]RadarrMovie, error) {
	response := make([]RadarrMovie, 1)
	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/movie/?apikey=%s", config.GetRadarBaseEndpoint(), config.GetRadarAPIKey()))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return response, nil
	}

	if resp.StatusCode == 200 {

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
