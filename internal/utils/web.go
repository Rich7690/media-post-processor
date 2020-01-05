package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var netClient = http.Client{
	Timeout: time.Second * 10,
}

func NetClient() *http.Client {
	return &netClient
}

func MakeWebRequest(method string, path string, body io.Reader) (*http.Response, []byte, error) {
	request, err := http.NewRequest(method, path, body)

	if err != nil {
		return nil, nil, err
	}
	resp, err := netClient.Do(request)

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	return resp, response, err
}

func GetRequest(path string) (*http.Response, []byte, error) {
	return MakeWebRequest("GET", path, nil)
}

func PostRequest(path string, body interface{}) (*http.Response, []byte, error) {
	value, err := json.Marshal(body)

	if err != nil {
		return nil, nil, err
	}

	buf := bytes.NewBuffer(value)

	return MakeWebRequest("POST", path, buf)
}
