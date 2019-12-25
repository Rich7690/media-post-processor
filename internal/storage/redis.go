package storage

import (
	"github.com/gomodule/redigo/redis"
	"media-web/internal/config"
)

// Make a redis pool
var RedisPool = redis.Pool{
	MaxActive: 10,
	MaxIdle:   10,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", config.GetRedisAddress())
	},
}
