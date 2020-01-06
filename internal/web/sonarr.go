package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"media-web/internal/config"
	"media-web/internal/utils"
)

func GetAllEpisodeFiles(seriesId int) ([]SonarrEpisodeFile, error) {
	response := make([]SonarrEpisodeFile, 1)
	path := fmt.Sprintf("%s/api/episodeFile/?seriesId=%d&apikey=%s", config.GetSonarrBaseEndpoint(), seriesId, config.GetSonarrAPIKey())
	resp, body, err := utils.GetRequest(path)

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

func GetAllSeries() ([]Series, error) {
	response := make([]Series, 1)
	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/series/?apikey=%s", config.GetSonarrBaseEndpoint(), config.GetSonarrAPIKey()))

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

func CheckSonarrCommand(id int) (*SonarrCommand, error) {
	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/command/%d?apikey=%s", config.GetSonarrBaseEndpoint(), id, config.GetSonarrAPIKey()))

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

func RescanSeries(id int64) (*SonarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanSeries"
	payload["seriesId"] = id

	resp, body, err := utils.PostRequest(fmt.Sprintf("%s/api/command/?apikey=%s", config.GetSonarrBaseEndpoint(), config.GetSonarrAPIKey()), payload)

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

func LookupTVEpisode(id int64) (*SonarrEpisodeFile, error) {
	resp, body, err := utils.GetRequest(fmt.Sprintf("%s/api/episodeFile/%d?apikey=%s", config.GetSonarrBaseEndpoint(), id, config.GetSonarrAPIKey()))

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
