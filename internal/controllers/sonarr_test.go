package controllers

import (
"bytes"
"encoding/json"
"errors"
"github.com/gin-gonic/gin"
"github.com/gocraft/work"
"github.com/golang/mock/gomock"
"io/ioutil"
"media-web/internal/constants"
"media-web/internal/helper"
"media-web/internal/web"
"media-web/internal/worker"
"net/http"
"net/http/httptest"
"testing"
)

func TestSonarrReturnsErrorForBadPayload(t *testing.T) {

	m := worker.WorkScheduler{
		EnqueueUnique: func(jobName string, args map[string]interface{}) (*work.Job, error) {
			return nil, nil
		},
	}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := bytes.NewBufferString("Not valid json")

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", body)

	helper.GetHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusBadRequest

		_, err := ioutil.ReadAll(w.Body)
		pageOK := err == nil

		return statusOK && pageOK
	})
}

func TestSonarrReturnsErrorForFailedEnqueue(t *testing.T) {

	m := worker.WorkScheduler{
		EnqueueUnique: func(jobName string, args map[string]interface{}) (*work.Job, error) {
			return nil, errors.New("Failed to enqueue job")
		},
	}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))

	helper.GetHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusInternalServerError

		_, err := ioutil.ReadAll(w.Body)
		pageOK := err == nil

		return statusOK && pageOK
	})
}

func TestSonarrEnqueuesJobForValidInput(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//m := mock_worker.NewMockWorkScheduler(ctrl)
	var count = 0
	var hitJobName = ""
	var hitArgs = make(map[string]interface{}, 1)
	m := worker.WorkScheduler{
		EnqueueUnique: func(jobName string, args map[string]interface{}) (*work.Job, error) {
			count++
			hitJobName = jobName
			hitArgs = args
			return &work.Job{ID: "mockJob"}, nil
		},
	}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))

	helper.GetHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		_, err := ioutil.ReadAll(w.Body)
		pageOK := err == nil

		match := hitJobName == constants.TranscodeJobType &&
			hitArgs[constants.TranscodeTypeKey] == constants.TV &&
			hitArgs[constants.EpisodeFileIdKey] == 0

		return statusOK && pageOK && match
	})
}

