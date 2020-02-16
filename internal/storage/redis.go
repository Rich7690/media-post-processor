package storage

import (
	"github.com/gomodule/redigo/redis"
	"media-web/internal/config"
	"time"
)

// Make a redis pool
var RedisPool = redis.Pool{
	MaxActive: 10,
	MaxIdle:   10,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.DialURL(config.GetConfig().RedisAddress.String(), redis.DialKeepAlive(5*time.Minute),
			redis.DialReadTimeout(5*time.Second), redis.DialConnectTimeout(5*time.Second))
	},
}
