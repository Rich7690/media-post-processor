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

func TestRescanMovieReturnsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := make(map[string]interface{})
		//nolint:errcheck
		json.NewDecoder(r.Body).Decode(&payload)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "RescanMovie", payload["name"])
		assert.EqualValues(t, 1, payload["movieId"])

		cmd := RadarrCommand{
			State: "complete",
			ID:    1,
		}
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	cmd, err := client.RescanMovie(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, cmd.ID)
}

func TestCheckRadarrCommandReturnsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, fmt.Sprintf("/api/command/%d", 1), r.URL.Path)

		cmd := RadarrCommand{
			State: "complete",
			ID:    1,
		}
		//nolint:errcheck
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	cmd, err := client.CheckRadarrCommand(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, cmd.ID)
}

func TestLookupMovieReturnsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, fmt.Sprintf("/api/movie/%d", 1), r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		cmd := RadarrMovie{
			ID: 1,
		}
		//nolint:errcheck
		json.NewEncoder(w).Encode(&cmd)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	movie, err := client.LookupMovie(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, movie.ID)
}

func TestGetAllMoviesReturnsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/movie", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		cmd := RadarrMovie{
			ID: 1,
		}
		mvs := []RadarrMovie{cmd}
		json.NewEncoder(w).Encode(&mvs)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	movie, err := client.GetAllMovies()

	assert.NoError(t, err)
	assert.Equal(t, 1, movie[0].ID)
}

func TestGetAllMoviesEmptyReturnsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/movie", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	movie, err := client.GetAllMovies()

	assert.NoError(t, err)
	assert.Equal(t, 0, len(movie))
}

func TestGetAllMoviesErrorReturnserror(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/movie", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "apikey=")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	client := RadarrClientImpl{
		webClient: utils.GetWebClient(parsed),
	}

	_, err := client.GetAllMovies()

	assert.Error(t, err)
}
