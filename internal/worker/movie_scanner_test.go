package worker

import (
	"errors"
	"github.com/gocraft/work"
	"media-web/internal/web"
	"testing"
)


func TestErrorFromScanner(t *testing.T) {

	mockErr := errors.New("Error!")

	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			return nil, mockErr
	}}, struct {
		EnqueueUnique func(jobName string, args map[string]interface{}) (*work.Job, error)
	}{EnqueueUnique: func(jobName string, args map[string]interface{}) (job *work.Job, e error) {
		return nil, nil
	}})

	if err != mockErr {
		t.Error("Errors don't match")
	}

}

func TestDoesNothingIfNoMovies(t *testing.T) {

	hitJob := false
	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			return make([]web.RadarrMovie, 1), nil
		}}, struct {
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
	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			movieList := make([]web.RadarrMovie, 1)
			movieList = append(movies, web.RadarrMovie{Downloaded: false})
			return movieList, nil
		}}, struct {
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
	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			movieList := make([]web.RadarrMovie, 1)
			movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mp4"}})
			return movieList, nil
		}}, struct {
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
	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			movieList := make([]web.RadarrMovie, 1)
			movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mkv"}})
			return movieList, nil
		}}, struct {
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
	err := ScanForMovies(struct{
		GetAllMovies func() ([]web.RadarrMovie, error)
	}{
		GetAllMovies: func() (movies []web.RadarrMovie, e error) {
			movieList := make([]web.RadarrMovie, 1)
			movieList = append(movies, web.RadarrMovie{Downloaded: true, MovieFile: web.MovieFile{RelativePath:"test.mkv"}})
			return movieList, nil
		}}, struct {
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
