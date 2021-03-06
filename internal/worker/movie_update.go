package worker

import (
	"media-web/internal/constants"
	"strconv"
	"strings"
	"time"

	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
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
			if strings.Contains(result.State, "complete") {
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
