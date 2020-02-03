package worker

import (
	"errors"
	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"media-web/internal/web"
	"testing"
)

func (c MockRadarr) CheckRadarrCommand (id int) (*web.RadarrCommand, error) {
	return c.checkRadarrCommand(id)
}
func (c MockRadarr) RescanMovie (id int64) (*web.RadarrCommand, error) {
	return c.rescanMovie(id)
}
func (c MockRadarr) LookupMovie (id int64) (*web.RadarrMovie, error) {
	return c.lookupMovie(id)
}
func (c MockRadarr) GetAllMovies () ([]web.RadarrMovie, error) {
	return c.getAllMovies()
}
func (c MockRadarr) GetMovieFilePath (id int64) (string, error) {
	return c.getMovieFilePath(id)
}

func TestErrorFromScanner(t *testing.T) {

	mockErr := errors.New("Error!")
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return nil, mockErr
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		return nil, nil
	}})

	assert.Error(t, err)

}

func TestDoesNothingIfNoMovies(t *testing.T) {

	hitJob := false
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		return make([]web.RadarrMovie, 1), nil
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		hitJob = true
		return nil, errors.New("Should not hit here")
	}})

	if err != nil {
		t.Error("Error returned")
	}
	if hitJob {
		t.Error("Should not have enqueued job")
	}

}

func TestDoesNothingIfNotDownloaded(t *testing.T) {

	hitJob := false
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movies, web.RadarrMovie{Downloaded: false})
		return movieList, nil
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		hitJob = true
		return nil, errors.New("Should not hit here")
	}})

	if err != nil {
		t.Error("Error returned")
	}
	if hitJob {
		t.Error("Should not have enqueued job")
	}

}

func TestSkipsIfAlreadyRightFormat(t *testing.T) {

	hitJob := false
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mp4"}})
		return movieList, nil
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		hitJob = true
		return nil, nil
	}})

	if err != nil {
		t.Error("Error returned")
	}
	if hitJob {
		t.Error("Should not have enqueued job")
	}

}

func TestEnqueueIfProperFormat(t *testing.T) {

	hitJob := false
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mkv"}})
		return movieList, nil
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		hitJob = true
		return nil, nil
	}})

	if err != nil {
		t.Error("Error returned")
	}
	if !hitJob {
		t.Error("Should not have enqueued job")
	}

}

func TestEnqueueAndIgnoresEnqueueError(t *testing.T) {

	hitJob := false
	mockClient := MockRadarr{}
	mockClient.getAllMovies = func() (movies []web.RadarrMovie, e error) {
		movieList := make([]web.RadarrMovie, 1)
		movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mkv"}})
		return movieList, nil
	}
	err := ScanForMovies(mockClient, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		hitJob = true
		return nil, errors.New("Error here!")
	}})

	if err != nil {
		t.Error("Error returned")
	}
	if !hitJob {
		t.Error("Should not have enqueued job")
	}

}
