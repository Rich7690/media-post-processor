package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"media-web/internal/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSonarrReturnsErrorForBadPayload(t *testing.T) {

	m := mockWorker{}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := bytes.NewBufferString("Not valid json")

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertExpectations(t)
}

func TestSonarrReturnsErrorForFailedEnqueue(t *testing.T) {

	m := mockWorker{}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	req, _ := http.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestSonarrEnqueuesJobForValidInput(t *testing.T) {

	m := mockWorker{}

	r := gin.Default()

	r.POST("/api/sonarr/webhook", GetSonarrWebhookHandler(m))

	body := web.SonarrWebhook{EventType: "Download"}

	payload, err := json.Marshal(body)

	if err != nil {
		t.Error("Failed to encode json")
	}

	req := httptest.NewRequest("POST", "/api/sonarr/webhook", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
