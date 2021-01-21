package worker

import (
	"media-web/internal/constants"
	"strconv"
	"strings"
	"time"

	"github.com/gocraft/work"
	"github.com/rs/zerolog/log"
)

func (c *WorkerContext) UpdateTVShow(job *work.Job) error {
	seriesId := job.ArgInt64(constants.SeriesIdKey)

	cmd, err := c.SonarrClient.RescanSeries(seriesId)

	if err != nil {
		log.Err(err).Msg("Error rescanning series")
		return err
	}

	for count := 0; count < 5; count++ {
		result, err := c.SonarrClient.CheckSonarrCommand(cmd.ID)

		if err == nil {
			if strings.Contains(result.State, "complete") {
				log.Info().Msg("Rescan complete for: " + strconv.Itoa(cmd.ID))
				return nil
			} else {
				log.Info().Msg("Rescan not complete yet for: " + strconv.Itoa(cmd.ID))
			}
		} else {
			log.Err(err).Msg("Error checking state of command: " + strconv.Itoa(cmd.ID))
		}

		c.Sleep(time.Second * 15)
	}

	return err
}
