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
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func CheckRadarrCommand(id int) (*RadarrCommand, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/command/%d?apikey=%s", config.GetRadarBaseEndpoint(), id, config.GetRadarAPIKey()), nil)

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
		response := RadarrCommand{}
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

func RescanMovie(id int64) (*RadarrCommand, error) {

	payload := make(map[string]interface{})

	payload["name"] = "RescanMovie"
	payload["movieId"] = id

	value, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(value)

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/command/?apikey=%s", config.GetRadarBaseEndpoint(), config.GetRadarAPIKey()), buf)

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
		fmt.Println(value)
		return nil, err
	}

	response := RadarrCommand{}

	err = json.Unmarshal(value, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func LookupMovie(id int64) (*RadarrMovie, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/movie/%d?apikey=%s", config.GetRadarBaseEndpoint(), id, config.GetRadarAPIKey()), nil)

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
		log.Println("Found movie")
		response := RadarrMovie{}
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
		return nil, errors.New("Failed find movie")
	}
}
