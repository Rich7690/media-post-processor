package worker

import (
	"errors"
	"media-web/internal/constants"
	"media-web/internal/web"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestErrorFromTVScanner(t *testing.T) {
	mockErr := errors.New("mock Error")
	mockClient := MockSonarr{}
	mockClient.getAllSeries = func() (series []web.Series, e error) {
		return nil, mockErr
	}
	// We'd fail with pointer errors if we called anything on here
	w := mockWorker{}
	ScanForTVShows(mockClient, &w)
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
	w.On("EnqueueUnique", constants.TranscodeJobType, map[string]interface{}{
		constants.TranscodeTypeKey: constants.TV,
		constants.EpisodeFileIDKey: 2,
	}).Once().Return(nil, nil)
	ScanForTVShows(mockClient, &w)
	assert.Equal(t, 1, inputSeries)
	w.AssertExpectations(t)
}
