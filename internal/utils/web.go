package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	path2 "path"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	NotFoundError = errors.New("http status not found returned")
)

type WebClient interface {
	GetRequest(url url.URL, path string, values url.Values, respObject interface{}) error
	PostRequest(url url.URL, path string, values url.Values, body interface{}, respObject interface{}) error
	MakeGetRequest(url url.URL, path string, values url.Values) (*http.Response, []byte, error)
	MakePostRequest(url url.URL, path string, values url.Values, body interface{}) (*http.Response, []byte, error)
}

type WebClientImpl struct {
	client http.Client
}

func (c WebClientImpl) PostRequest(url url.URL, path string, values url.Values, body interface{}, respObject interface{}) error {
	resp, respBytes, err := makePostRequest(url, path, values, body)

	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code from server: " + strconv.Itoa(resp.StatusCode))
	}
	if respObject != nil {
		err = json.Unmarshal(respBytes, respObject)
	}

	return err
}

func (c WebClientImpl) GetRequest(url url.URL, path string, values url.Values, respObject interface{}) error {
	resp, body, err := makeGetRequest(url, path, values)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return NotFoundError
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code from server: " + strconv.Itoa(resp.StatusCode))
	}
	if respObject != nil {
		err = json.Unmarshal(body, respObject)
	}

	return err
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
