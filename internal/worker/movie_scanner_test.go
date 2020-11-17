package worker

import (
	"errors"
	"media-web/internal/constants"
	"media-web/internal/web"
	"testing"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (c MockRadarr) CheckRadarrCommand(id int) (*web.RadarrCommand, error) {
	return c.checkRadarrCommand(id)
}
func (c MockRadarr) RescanMovie(id int64) (*web.RadarrCommand, error) {
	return c.rescanMovie(id)
}
func (c MockRadarr) LookupMovie(id int64) (*web.RadarrMovie, error) {
	return c.lookupMovie(id)
}
func (c MockRadarr) GetAllMovies() ([]web.RadarrMovie, error) {
	return c.getAllMovies()
}
func (c MockRadarr) GetMovieFilePath(id int64) (string, error) {
	return c.getMovieFilePath(id)
}

type mockWorker struct {
	mock.Mock
}

func (m *mockWorker) EnqueueUnique(jobName string, args map[string]interface{}) (*work.Job, error) {
	resp := m.Called(jobName, args)

	arg := resp.Get(0)

	job, ok := arg.(*work.Job)

	if ok {
		return job, resp.Error(1)
	}

	return nil, resp.Error(1)
}

func TestErrorFromScanner(t *testing.T) {

	mockErr := errors.New("Error!")
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return nil, mockErr
	}
	w := mockWorker{}
	w.On("EnqueueUnique").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	assert.Error(t, err)
	w.AssertNotCalled(t, "EnqueueUnique")

}

func TestDoesNothingIfNoMovies(t *testing.T) {

	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return make([]web.RadarrMovie, 1), nil
	}
	w := mockWorker{}
	w.On("EnqueueUnique").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueUnique")
}

func TestDoesNothingIfNotDownloaded(t *testing.T) {

	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movieList, web.RadarrMovie{Downloaded: false})
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueUnique").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueUnique")
}

func TestSkipsIfAlreadyRightFormat(t *testing.T) {

	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mp4"}})
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueUnique").Times(0).Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertNotCalled(t, "EnqueueUnique")
}

func TestEnqueueIfProperFormat(t *testing.T) {

	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := []web.RadarrMovie{web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mkv"}}}
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueUnique", constants.TranscodeJobType, map[string]interface{}{constants.MovieIdKey: 0,
		constants.TranscodeTypeKey: constants.Movie}).Once().Return(nil, nil)
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertExpectations(t)

}

func TestEnqueueAndIgnoresEnqueueError(t *testing.T) {

	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := []web.RadarrMovie{web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath: "test.mkv"}}}
		return movieList, nil
	}
	w := mockWorker{}
	w.On("EnqueueUnique", mock.Anything, mock.Anything).Once().Return(nil, errors.New("boom"))
	scanner := NewMovieScanner(mockClient, &w)
	err := scanner.ScanForMovies()

	if err != nil {
		t.Error("Error returned")
	}
	w.AssertExpectations(t)

}
