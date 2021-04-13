package worker

import (
	"context"
	"errors"
	"media-web/internal/storage"
	"media-web/internal/web"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorker struct {
	mock.Mock
}

func (m *mockWorker) EnqueueJob(ctx context.Context, job *storage.TranscodeJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *mockWorker) DequeueJob(ctx context.Context, work func(ctx context.Context, job storage.TranscodeJob) error) error {
	return m.Called(ctx, work).Error(0)
}
func (m *mockWorker) HandleErrored(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestErrorFromScanner(t *testing.T) {
	mockErr := errors.New("error")
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return nil, mockErr
	}
	w := mockWorker{}
	w.On("EnqueueJob").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	assert.Error(t, err)
	w.AssertNotCalled(t, "EnqueueJob")
}

func TestDoesNothingIfNoMovies(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return make([]web.RadarrMovie, 1), nil
	}
	w := mockWorker{}
	w.On("EnqueueJob").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueJob")
}

func TestDoesNothingIfNotDownloaded(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movieList, web.RadarrMovie{Downloaded: false})
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueJob").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueJob")
}

func TestSkipsIfAlreadyRightFormat(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mp4"}})
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueJob").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueJob")
}

func TestEnqueueIfProperFormat(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := []web.RadarrMovie{{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mkv"}}}
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueJob", mock.Anything, mock.Anything).Once().Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertExpectations(t)
}

func TestEnqueueAndIgnoresEnqueueError(t *testing.T) {
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := []web.RadarrMovie{{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mkv"}}}
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueJob", mock.Anything, mock.Anything).Once().Return(nil, errors.New("boom"))
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies(context.Background())

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertExpectations(t)
}
