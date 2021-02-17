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
	ErrNotFound = errors.New("http status not found returned")
)

type WebClient interface {
	GetRequest(path string, values url.Values, respObject interface{}) error
	PostRequest(path string, values url.Values, body interface{}, respObject interface{}) error
	MakeGetRequest(path string, values url.Values) (*http.Response, []byte, error)
	MakePostRequest(path string, values url.Values, body interface{}) (*http.Response, []byte, error)
}

type WebClientImpl struct {
	client http.Client
	base   *url.URL
}

func (c WebClientImpl) PostRequest(path string, values url.Values, body, respObject interface{}) error {
	resp, respBytes, err := makePostRequest(c.base, path, values, body)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code from server: " + strconv.Itoa(resp.StatusCode))
	}
	if respObject != nil {
		err = json.Unmarshal(respBytes, respObject)
	}

	return err
}

func (c WebClientImpl) GetRequest(path string, values url.Values, respObject interface{}) error {
	resp, body, err := makeGetRequest(c.base, path, values)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code from server: " + strconv.Itoa(resp.StatusCode))
	}
	if respObject != nil {
		err = json.Unmarshal(body, respObject)
	}

	return err
}

func GetWebClient(base *url.URL) WebClient {
	return WebClientImpl{client: netClient, base: base}
}

func (c WebClientImpl) MakeGetRequest(path string, values url.Values) (*http.Response, []byte, error) {
	return makeGetRequest(c.base, path, values)
}

func (c WebClientImpl) MakePostRequest(path string, values url.Values, requestBody interface{}) (*http.Response, []byte, error) {
	return makePostRequest(c.base, path, values, requestBody)
}

var netClient = http.Client{
	Timeout: time.Second * 10,
}

func makePostRequest(base *url.URL, path string, values url.Values, body interface{}) (*http.Response, []byte, error) {
	log.Trace().Str("base", base.String()).Str("path", path).Msg("preparing post")
	value, err := json.Marshal(body)

	if err != nil {
		return nil, nil, err
	}
	baseURL := *base
	buf := bytes.NewBuffer(value)
	finalPath := path2.Join(baseURL.Path, path)
	baseURL.Path = finalPath
	currentValues := baseURL.Query()
	for k, v := range values {
		for _, value := range v {
			currentValues.Add(k, value)
		}
	}
	baseURL.RawQuery = currentValues.Encode()
	log.Trace().Str("url", baseURL.String()).Msg("Making POST request")
	//nolint:noctx
	resp, err := netClient.Post(baseURL.String(), "application/json", buf)

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	return resp, response, err
}

func makeGetRequest(baseURL *url.URL, path string, values url.Values) (*http.Response, []byte, error) {
	base := *baseURL
	log.Trace().Str("base", base.String()).Str("path", path).Msg("preparing get")
	finalPath := path2.Join(base.Path, path)
	base.Path = finalPath
	currentValues := base.Query()
	for k, v := range values {
		for _, value := range v {
			currentValues.Add(k, value)
		}
	}
	base.RawQuery = currentValues.Encode()
	log.Trace().Str("url", base.String()).Msg("Making GET request")
	//nolint:noctx
	resp, err := netClient.Get(base.String())

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	return resp, response, err
}
