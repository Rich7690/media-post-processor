package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"media-web/internal/config"
	"net/http"
)

func GetAllEpisodeFiles(seriesId int) ([]SonarrEpisodeFile, error) {
	response := make([]SonarrEpisodeFile, 1)
	path := fmt.Sprintf("%s/api/episodeFile/?seriesId=%d&apikey=%s", config.GetSonarrBaseEndpoint(), seriesId, config.GetSonarrAPIKey())
	request, err := http.NewRequest("GET", path, nil)

	if err != nil {
		return nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return response, nil
	}

	if resp.StatusCode == 200 {

		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Println("Got error parsing json: ", err.Error())
			log.Println(string(body))
			return nil, err
		}

		return response, nil

	} else {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		return nil, errors.New("Failed to get episode files")
	}
}

func GetAllSeries() ([]Series, error) {
	response := make([]Series, 1)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/series/?apikey=%s", config.GetSonarrBaseEndpoint(), config.GetSonarrAPIKey()), nil)

	if err != nil {
		return nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return response, nil
	}

	if resp.StatusCode == 200 {

		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Println("Got error parsing json: ", err.Error())
			log.Println(string(body))
			return nil, err
		}

		return response, nil

	} else {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		return nil, errors.New("Failed to get series")
	}
}

func CheckSonarrCommand(id int) (*SonarrCommand, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/command/%d?apikey=%s", config.GetSonarrBaseEndpoint(), id, config.GetSonarrAPIKey()), nil)

	if err != nil {
		return nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		response := SonarrCommand{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Println("Got error parsing json: ", err.Error())
			log.Println(string(body))
			return nil, err
		}

		return &response, nil

	} else {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		return nil, errors.New("Failed to check command")
	}
}

func RescanSeries(id int64) (*SonarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanSeries"
	payload["seriesId"] = id

	value, err := json.Marshal(payload)

	buf := bytes.NewBuffer(value)

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/command/?apikey=%s", config.GetSonarrBaseEndpoint(), config.GetSonarrAPIKey()), buf)

	if err != nil {
		return nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	value, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	response := SonarrCommand{}

	err = json.Unmarshal(value, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func LookupTVEpisode(id int64) (*SonarrEpisodeFile, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/episodeFile/%d?apikey=%s", config.GetSonarrBaseEndpoint(), id, config.GetSonarrAPIKey()), nil)

	if err != nil {
		return nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode == 200 {
		log.Println("Found episode file")
		response := SonarrEpisodeFile{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			log.Println("Got error parsing json: ", err.Error())
			log.Println(string(body))
			return nil, err
		}

		return &response, nil

	} else {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		return nil, errors.New("Failed find episode file")
	}
}
