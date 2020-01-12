package storage

import (
	red "github.com/go-redis/redis/v7"
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

var RedisClient = red.NewClient(&red.Options{
	Addr:     config.GetRedisAddress(),
	Password: "", // no password set
	DB:       0,  // use default DB

})
