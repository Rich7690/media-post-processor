package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"media-web/internal/config"
	"media-web/internal/transcode"
	"media-web/internal/web"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type RadarrCache interface {
	web.RadarrClient
	SyncMovies() error
}

type RadarrCacheImpl struct {
	client web.RadarrClient
	redis  *redis.Client
}

func GetRadarrCache() RadarrCache {
	redisAddr := config.GetConfig().RedisAddress
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr.Host,
		DB:   0, // use default DB
	})

	return RadarrCacheImpl{
		client: web.GetRadarrClient(),
		redis:  rdb,
	}
}

func (r RadarrCacheImpl) SyncMovies() error {
	results, err := r.client.GetAllMovies()
	if err != nil {
		return errors.Wrap(err, "Failed to query radarr for movies")
	}

	for i := range results {
		buf, lerr := json.Marshal(&results[i])
		if lerr != nil {
			return errors.Wrap(lerr, "Failed to encode movie")
		}
		lerr = r.redis.Set(context.Background(), fmt.Sprintf("radarr:movies:%d", results[i].ID), string(buf), 0).Err()
		if lerr != nil {
			log.Err(lerr).Msg("Failed to cache movies result")
		}
	}
	return nil
}

func (r RadarrCacheImpl) CheckRadarrCommand(id int) (*web.RadarrCommand, error) {
	panic("not implemented") // TODO: Implement
}

func (r RadarrCacheImpl) RescanMovie(id int64) (*web.RadarrCommand, error) {
	panic("not implemented") // TODO: Implement
}

func (r RadarrCacheImpl) LookupMovie(id int64) (*web.RadarrMovie, error) {
	result, err := r.redis.Get(context.Background(), fmt.Sprintf("radarr:movies:%d", id)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString(result)
	movie := web.RadarrMovie{}
	err = json.NewDecoder(buf).Decode(&movie)
	return &movie, err
}

func (r RadarrCacheImpl) GetAllMovies() ([]web.RadarrMovie, error) {
	result := make([]web.RadarrMovie, 0)
	var curs uint64 = 0

	for {
		keys, cur, err := r.redis.Scan(context.Background(), curs, "radarr:movies:*", 100).Result()
		if err != nil {
			return result, err
		}
		if len(keys) > 0 {
			encoded, err := r.redis.MGet(context.Background(), keys...).Result()
			if err != nil {
				return result, err
			}
			for i := range encoded {
				movie := web.RadarrMovie{}
				err = json.Unmarshal([]byte(encoded[i].(string)), &movie)
				if err != nil {
					return result, err
				}
				result = append(result, movie)
			}
		}
		if cur == 0 {
			break
		}
		curs = cur
	}

	return result, nil
}

func (r RadarrCacheImpl) GetMovieFilePath(id int64) (transcode.VideoFile, error) {
	panic("not implemented") // TODO: Implement
}

func (r RadarrCacheImpl) ScanForMissingMovies() (*web.RadarrCommand, error) {
	panic("not implemented") // TODO: Implement
}
