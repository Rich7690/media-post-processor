package worker

import (
	"context"
	"errors"
	"media-web/internal/web"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestErrorFromTVScanner(t *testing.T) {
	mockErr := errors.New("mock Error")
	mockClient := MockSonarr{}
	mockClient.getAllSeries = func() (series []web.Series, e error) {
		return nil, mockErr
	}
	// We'd fail with pointer errors if we called anything on here
	w := mockWorker{}
	ScanForTVShows(context.Background(), mockClient, &w)
	w.AssertExpectations(t)
}

func TestSkipIfUnmatchedExtension(t *testing.T) {
	mockSeries := make([]web.Series, 0)
	mockSeries = append(mockSeries, web.Series{
		Title: "TestTitle",
		ID:    1,
	})
	episodes := make([]web.SonarrEpisodeFile, 0)
	episodes = append(episodes, web.SonarrEpisodeFile{
		Path: "test.mkv",
		ID:   2,
	})
	var inputSeries = -1
	mockClient := MockSonarr{
		getAllSeries: func() ([]web.Series, error) {
			return mockSeries, nil
		},
		getAllEpisodeFiles: func(seriesId int) ([]web.SonarrEpisodeFile, error) {
			inputSeries = seriesId
			return episodes, nil
		},
	}
	w := mockWorker{}
	w.On("EnqueueJob", mock.Anything, mock.Anything).Once().Return(nil)
	ScanForTVShows(context.Background(), mockClient, &w)
	assert.Equal(t, 1, inputSeries)
	w.AssertExpectations(t)
}
