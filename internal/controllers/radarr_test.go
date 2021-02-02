package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"media-web/internal/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorker struct {
	mock.Mock
}

func (m *mockWorker) EnqueueUnique(jobName string, args map[string]interface{}) (*work.Job, error) {
	resp := m.Called(jobName, args)

	arg := resp.Get(0)

	job, ok := arg.(*work.Job)

	if ok {
		return job, resp.Error(1)
	}

	return nil, resp.Error(1)
}

func TestReturnsErrorForBadPayload(t *testing.T) {
	m := &mockWorker{}

	body := bytes.NewBufferString("Not valid json")

	req, _ := http.NewRequest("POST", "/api/radarr/webhook", body)
	w := httptest.NewRecorder()
	GetRadarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertExpectations(t)
}

func TestReturnsErrorForFailedEnqueue(t *testing.T) {
	m := &mockWorker{}

	body := web.RadarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}
	m.On("EnqueueUnique", mock.Anything, mock.Anything).Return(&work.Job{}, errors.New("boom"))
	req, _ := http.NewRequest("POST", "/api/radarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	GetRadarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestEnqueuesJobForValidInput(t *testing.T) {
	m := &mockWorker{}

	movie := web.Movie{ID: 1}
	body := web.RadarrWebhook{EventType: "Download", Movie: movie}
	job := work.Job{ID: "foo"}

	m.On("EnqueueUnique", mock.Anything, mock.Anything).Return(&job, nil)

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	req, _ := http.NewRequest("POST", "/api/radarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	GetRadarrWebhookHandler(m)(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
