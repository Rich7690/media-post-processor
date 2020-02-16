package utils

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/url"
	path2 "path"
	"time"
)

type WebClient interface {
	MakeGetRequest(url url.URL, path string, values url.Values) (*http.Response, []byte, error)
	MakePostRequest(url url.URL, path string, values url.Values, body interface{}) (*http.Response, []byte, error)
}

type WebClientImpl struct {
	client http.Client
}

func GetWebClient() WebClient {
	return WebClientImpl{client: netClient}
}

func (c WebClientImpl) MakeGetRequest(url url.URL, path string, values url.Values) (*http.Response, []byte, error) {
	return makeGetRequest(url, path, values)
}

func (c WebClientImpl) MakePostRequest(url url.URL, path string, values url.Values, body interface{}) (*http.Response, []byte, error) {
	return makePostRequest(url, path, values, body)
}

var netClient = http.Client{
	Timeout: time.Second * 10,
}

func makePostRequest(url url.URL, path string, values url.Values, body interface{}) (*http.Response, []byte, error) {
	value, err := json.Marshal(body)

	if err != nil {
		return nil, nil, err
	}

	buf := bytes.NewBuffer(value)
	finalPath := path2.Join(url.Path, path)
	url.RawPath = finalPath
	url.Path = finalPath
	currentValues := url.Query()
	for k, v := range values {
		for _, value := range v {
			currentValues.Add(k, value)
		}
	}
	url.RawQuery = currentValues.Encode()
	log.Trace().Str("url", url.String()).Msg("Making POST request")
	resp, err := netClient.Post(url.String(), "application/json", buf)

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	return resp, response, err
}

func makeGetRequest(url url.URL, path string, values url.Values) (*http.Response, []byte, error) {
	finalPath := path2.Join(url.Path, path)
	url.RawPath = finalPath
	url.Path = finalPath
	currentValues := url.Query()
	for k, v := range values {
		for _, value := range v {
			currentValues.Add(k, value)
		}
	}
	url.RawQuery = currentValues.Encode()
	log.Trace().Str("url", url.String()).Msg("Making GET request")
	resp, err := netClient.Get(url.String())

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	return resp, response, err
}
