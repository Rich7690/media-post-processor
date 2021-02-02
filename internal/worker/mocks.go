package worker

import (
	"media-web/internal/transcode"
	"media-web/internal/web"
)

type MockSonarr struct {
	getAllSeries       func() ([]web.Series, error)
	getAllEpisodeFiles func(seriesId int) ([]web.SonarrEpisodeFile, error)
	checkSonarrCommand func(id int) (*web.SonarrCommand, error)
	rescanSeries       func(id *int64) (*web.SonarrCommand, error)
	lookupTVEpisode    func(id int64) (*web.SonarrEpisodeFile, error)
	getEpisodeFilePath func(id int64) (transcode.VideoFile, int, error)
}

type MockRadarr struct {
	checkRadarrCommand func(id int) (*web.RadarrCommand, error)
	rescanMovie        func(id int64) (*web.RadarrCommand, error)
	lookupMovie        func(id int64) (*web.RadarrMovie, error)
	getAllMovies       func() ([]web.RadarrMovie, error)
	getMovieFilePath   func(id int64) (transcode.VideoFile, error)
}

func (m MockSonarr) GetAllEpisodeFiles(seriesID int) ([]web.SonarrEpisodeFile, error) {
	return m.getAllEpisodeFiles(seriesID)
}

func (m MockSonarr) CheckSonarrCommand(id int) (*web.SonarrCommand, error) {
	return m.checkSonarrCommand(id)
}

func (m MockSonarr) RescanSeries(id *int64) (*web.SonarrCommand, error) {
	return m.rescanSeries(id)
}

func (m MockSonarr) LookupTVEpisode(id int64) (*web.SonarrEpisodeFile, error) {
	return m.lookupTVEpisode(id)
}

func (m MockSonarr) GetEpisodeFilePath(id int64) (transcode.VideoFile, int, error) {
	return m.getEpisodeFilePath(id)
}

func (m MockSonarr) GetAllSeries() ([]web.Series, error) {
	return m.getAllSeries()
}

func (c MockRadarr) ScanForMissingMovies() (*web.RadarrCommand, error) {
	panic("implement me")
}

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
func (c MockRadarr) GetMovieFilePath(id int64) (transcode.VideoFile, error) {
	return c.getMovieFilePath(id)
}
