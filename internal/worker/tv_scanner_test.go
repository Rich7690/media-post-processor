package worker

import (
	"errors"
	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"media-web/internal/constants"
	"media-web/internal/web"
	"testing"
)


func (m MockSonarr) GetAllEpisodeFiles(seriesId int) ([]web.SonarrEpisodeFile, error) {
	return m.getAllEpisodeFiles(seriesId)
}

func (m MockSonarr) CheckSonarrCommand(id int) (*web.SonarrCommand, error) {
	return m.checkSonarrCommand(id)
}

func (m MockSonarr) RescanSeries(id int64) (*web.SonarrCommand, error) {
	return m.rescanSeries(id)
}

func (m MockSonarr) LookupTVEpisode(id int64) (*web.SonarrEpisodeFile, error) {
	return m.lookupTVEpisode(id)
}

func (m MockSonarr) GetEpisodeFilePath(id int64) (string, int, error) {
	return m.getEpisodeFilePath(id)
}

func (m MockSonarr) GetAllSeries() ([]web.Series, error) {
	return m.getAllSeries()
}

func TestErrorFromTVScanner(t *testing.T) {

	mockErr := errors.New("mock Error")
	mockClient :=  MockSonarr{}
	mockClient.getAllSeries = func() (series []web.Series, e error) {
		return nil, mockErr
	}
	// We'd fail with pointer errors if we called anything on here
	ScanForTVShows(mockClient, WorkScheduler{})
}

func TestSkipIfUnmatchedExtension(t *testing.T) {

	mockSeries := make([]web.Series, 1)
	mockSeries = append(mockSeries, web.Series{
		Title: "TestTitle",
		ID:    1,
	})
	episodes := make([]web.SonarrEpisodeFile, 1)
	episodes = append(episodes, web.SonarrEpisodeFile{
		Path: "test.mkv",
		ID:   2,
	})
	inputSeries := -1
	mockClient := MockSonarr{
		getAllSeries: func() ([]web.Series, error) {
			return mockSeries, nil
		},
		getAllEpisodeFiles: func(seriesId int) ([]web.SonarrEpisodeFile, error) {
			inputSeries = seriesId
			return episodes, nil
		},
	}
	inputJob := ""
	var inputArgs map[string]interface{}
	mockScheduler := WorkScheduler{
		EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
			inputJob = jobName
			inputArgs = args
			return nil, nil
		}}
	ScanForTVShows(mockClient, mockScheduler)

	assert.EqualValues(t, constants.TranscodeJobType, inputJob)
	assert.EqualValues(t, 1, inputSeries)
	assert.EqualValues(t, work.Q{
		constants.TranscodeTypeKey: constants.TV,
		constants.EpisodeFileIdKey: 2,
	}, inputArgs)
}
