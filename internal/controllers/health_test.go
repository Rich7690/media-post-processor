package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"media-web/internal/helper"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	r := gin.Default()

	r.GET("/health", HealthHandler)

	req, _ := http.NewRequest("GET", "/health", nil)

	helper.GetHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		_, err := ioutil.ReadAll(w.Body)
		pageOK := err == nil

		return statusOK && pageOK
	})
}
