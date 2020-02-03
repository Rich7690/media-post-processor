package worker

import "media-web/internal/web"

type MockSonarr struct {
	getAllSeries       func() ([]web.Series, error)
	getAllEpisodeFiles func(seriesId int) ([]web.SonarrEpisodeFile, error)
	checkSonarrCommand func(id int) (*web.SonarrCommand, error)
	rescanSeries       func(id int64) (*web.SonarrCommand, error)
	lookupTVEpisode    func(id int64) (*web.SonarrEpisodeFile, error)
	getEpisodeFilePath func(id int64) (string, int, error)
}

type MockRadarr struct {
	checkRadarrCommand func(id int) (*web.RadarrCommand, error)
	rescanMovie func(id int64) (*web.RadarrCommand, error)
	lookupMovie func(id int64) (*web.RadarrMovie, error)
	getAllMovies func() ([]web.RadarrMovie, error)
	getMovieFilePath func(id int64) (string, error)
}
