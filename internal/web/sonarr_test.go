package web

import (
	"encoding/json"
	"fmt"
	"media-web/internal/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRescanTVReturnsSuccess(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := make(map[string]interface{})

		json.NewDecoder(r.Body).Decode(&payload)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "RescanSeries", payload["name"])
		assert.EqualValues(t, 1, payload["seriesId"])

		cmd := SonarrCommand{
			State: "complete",
			ID:    1,
		}
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: *parsed,
	}

	cmd, err := client.RescanSeries(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, cmd.ID)
}

func TestCheckSonarrCommandReturnsSuccess(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, fmt.Sprintf("/api/command/%d", 1), r.URL.Path)

		cmd := SonarrCommand{
			State: "complete",
			ID:    1,
		}
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: *parsed,
	}

	cmd, err := client.CheckSonarrCommand(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, cmd.ID)
}

func TestGetAllSeriesReturnsSuccess(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/series", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		cmd := Series{
			ID: 1,
		}
		mvs := []Series{cmd}
		json.NewEncoder(w).Encode(&mvs)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: *parsed,
	}

	series, err := client.GetAllSeries()

	assert.NoError(t, err)
	assert.Equal(t, 1, series[0].ID)
}

func TestGetAllEpisodeFilesReturnsSuccess(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/episodeFile", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		assert.Equal(t, "1", r.URL.Query().Get("seriesId"))
		cmd := SonarrEpisodeFile{
			ID: 2,
		}
		mvs := []SonarrEpisodeFile{cmd}
		json.NewEncoder(w).Encode(&mvs)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: *parsed,
	}

	series, err := client.GetAllEpisodeFiles(1)

	assert.NoError(t, err)
	assert.Equal(t, 2, series[0].ID)
}

func TestLookupTVEpisodeReturnsSuccess(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, fmt.Sprintf("/api/episodeFile/%d", 1), r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		cmd := SonarrEpisodeFile{
			ID: 1,
		}
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := SonarrClientImpl{
		webClient:          utils.GetWebClient(),
		BaseSonarrEndpoint: *parsed,
	}

	episodeFile, err := client.LookupTVEpisode(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, episodeFile.ID)
}
