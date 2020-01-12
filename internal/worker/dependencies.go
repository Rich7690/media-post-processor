package worker

import "media-web/internal/web"

var TVScannerImpl = TVScanner{
	GetAllSeries:       web.GetAllSeries,
	GetAllEpisodeFiles: web.GetAllEpisodeFiles,
}

var MovieScannerImpl = MovieScanner{
	GetAllMovies:web.GetAllMovies,
}
