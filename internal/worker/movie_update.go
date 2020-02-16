package worker

import (
	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
	"media-web/internal/constants"
	"strconv"
	"strings"
	"time"
)

func (c *WorkerContext) UpdateMovie(job *work.Job) error {
	movieId := job.ArgInt64(constants.MovieIdKey)

	cmd, err := c.RadarrClient.RescanMovie(movieId)

	if err != nil {
		log.Err(err).Msg("Error rescanning movie: " + strconv.Itoa(int(movieId)))
		return err
	}

	for count := 0; count < 5; count++ {
		result, err := c.RadarrClient.CheckRadarrCommand(cmd.ID)

		if err == nil {
			if strings.Index(result.State, "complete") != -1 {
				log.Info().Msgf("Rescan complete for: %d", cmd.ID)
				return nil
			} else {
				log.Info().Msgf("Rescan not complete yet for: %d", cmd.ID)
			}
		} else {
			log.Err(err).Msg("Error checking status of command")
		}

		c.Sleep(time.Second * 15)
	}

	return nil
}
