package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"media-web/internal/constants"
	"media-web/internal/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocraft/work"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestSonarrReturnsErrorForBadPayload(t *testing.T) {
	m := mockWorker{}

	body := bytes.NewBufferString("Not valid json")

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", body)
	w := httptest.NewRecorder()
	GetSonarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertExpectations(t)
}

func TestSonarrReturnsErrorForFailedEnqueue(t *testing.T) {
	m := mockWorker{}

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	m.On("EnqueueUnique", constants.TranscodeJobType, mock.Anything).Return(&work.Job{}, errors.New("boom"))
	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	GetSonarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestSonarrEnqueuesJobForValidInput(t *testing.T) {
	m := mockWorker{}

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	m.On("EnqueueUnique", constants.TranscodeJobType, mock.Anything).Return(&work.Job{ID: "blah"}, nil)
	req := httptest.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	GetSonarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
