package worker

import (
	"errors"
	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"media-web/internal/constants"
	"media-web/internal/web"
	"testing"
	"time"
)

func TestRescanSeriesErrorReturns(t *testing.T) {
	mockClient := MockSonarr{}
	mockClient.rescanSeries = func(id int64) (command *web.SonarrCommand, e error) {
		return nil, errors.New("test error")
	}
	context := WorkerContext{
		SonarrClient:  mockClient,
	}

	err := context.UpdateTVShow(&work.Job{Args: map[string]interface{}{constants.SeriesIdKey: 1}})

	assert.Error(t, err)
}

func TestRescanNoErrorOnNonCompleteAfterAllTries(t *testing.T) {
	mockClient := MockSonarr{}
	callRescan := false
	mockClient.rescanSeries = func(id int64) (*web.SonarrCommand, error) {
		callRescan = true
		return &web.SonarrCommand{ID:1}, nil
	}
	callCheck := false
	mockClient.checkSonarrCommand = func(id int) (*web.SonarrCommand, error) {
		callCheck = true
		return &web.SonarrCommand{ID:1, State:""}, nil
	}
	context := WorkerContext{
		SonarrClient:  mockClient,
		Sleep: func(d time.Duration) {},
	}

	err := context.UpdateTVShow(&work.Job{Args: map[string]interface{}{constants.SeriesIdKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}

func TestRescanNoErrorOnComplete(t *testing.T) {
	mockClient := MockSonarr{}
	callRescan := false
	mockClient.rescanSeries = func(id int64) (*web.SonarrCommand, error) {
		callRescan = true
		return &web.SonarrCommand{ID:1}, nil
	}
	callCheck := false
	mockClient.checkSonarrCommand = func(id int) (*web.SonarrCommand, error) {
		callCheck = true
		return &web.SonarrCommand{ID:1, State:"complete"}, nil
	}
	context := WorkerContext{
		SonarrClient:  mockClient,
		Sleep: func(d time.Duration) {},
	}

	err := context.UpdateTVShow(&work.Job{Args: map[string]interface{}{constants.SeriesIdKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}

func TestRescanNoErrorCheckCommand(t *testing.T) {
	mockClient := MockSonarr{}
	callRescan := false
	mockClient.rescanSeries = func(id int64) (*web.SonarrCommand, error) {
		callRescan = true
		return &web.SonarrCommand{ID:1}, nil
	}
	callCheck := false
	mockClient.checkSonarrCommand = func(id int) (*web.SonarrCommand, error) {
		callCheck = true
		return &web.SonarrCommand{ID:1, State:""}, errors.New("test error")
	}
	context := WorkerContext{
		SonarrClient:  mockClient,
		Sleep: func(d time.Duration) {},
	}

	err := context.UpdateTVShow(&work.Job{Args: map[string]interface{}{constants.SeriesIdKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}
