package worker

import (
	"errors"
	"media-web/internal/constants"
	"media-web/internal/web"
	"testing"
	"time"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
)

func TestRadarrRescanSeriesErrorReturns(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.rescanMovie = func(id int64) (command *web.RadarrCommand, e error) {
		return nil, errors.New("test error")
	}
	context := WorkerContext{
		RadarrClient: mockClient,
	}

	err := context.UpdateMovie(&work.Job{Args: map[string]interface{}{constants.SeriesIDKey: 1}})

	assert.Error(t, err)
}

func TestRadarrRescanNoErrorOnNonCompleteAfterAllTries(t *testing.T) {
	mockClient := MockRadarr{}
	callRescan := false
	mockClient.rescanMovie = func(id int64) (*web.RadarrCommand, error) {
		callRescan = true
		return &web.RadarrCommand{ID: 1}, nil
	}
	callCheck := false
	mockClient.checkRadarrCommand = func(id int) (*web.RadarrCommand, error) {
		callCheck = true
		return &web.RadarrCommand{ID: 1, State: ""}, nil
	}
	context := WorkerContext{
		RadarrClient: mockClient,
		Sleep:        func(d time.Duration) {},
	}

	err := context.UpdateMovie(&work.Job{Args: map[string]interface{}{constants.MovieIDKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}

func TestRadarrRescanNoErrorOnComplete(t *testing.T) {
	mockClient := MockRadarr{}
	callRescan := false
	mockClient.rescanMovie = func(id int64) (*web.RadarrCommand, error) {
		callRescan = true
		return &web.RadarrCommand{ID: 1}, nil
	}
	callCheck := false
	mockClient.checkRadarrCommand = func(id int) (*web.RadarrCommand, error) {
		callCheck = true
		return &web.RadarrCommand{ID: 1, State: "complete"}, nil
	}
	context := WorkerContext{
		RadarrClient: mockClient,
		Sleep:        func(d time.Duration) {},
	}

	err := context.UpdateMovie(&work.Job{Args: map[string]interface{}{constants.MovieIDKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}

func TestRadarrRescanNoErrorCheckCommand(t *testing.T) {
	mockClient := MockRadarr{}
	callRescan := false
	mockClient.rescanMovie = func(id int64) (*web.RadarrCommand, error) {
		callRescan = true
		return &web.RadarrCommand{ID: 1}, nil
	}
	callCheck := false
	mockClient.checkRadarrCommand = func(id int) (*web.RadarrCommand, error) {
		callCheck = true
		return &web.RadarrCommand{ID: 1, State: ""}, errors.New("test error")
	}
	context := WorkerContext{
		RadarrClient: mockClient,
		Sleep:        func(d time.Duration) {},
	}

	err := context.UpdateMovie(&work.Job{Args: map[string]interface{}{constants.MovieIDKey: 1}})

	assert.True(t, callCheck)
	assert.True(t, callRescan)
	assert.NoError(t, err)
}
