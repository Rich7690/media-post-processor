package storage

import (
	"media-web/internal/config"
	"time"

	"github.com/bsm/redislock"
	redis "github.com/go-redis/redis/v8"
	redigo "github.com/gomodule/redigo/redis"
)

// RedisPool makes a redis pool
var RedisPool = redigo.Pool{
	MaxActive:   10,
	MaxIdle:     10,
	IdleTimeout: 10 * time.Minute,
	Wait:        true,
	Dial: func() (redigo.Conn, error) {
		return redigo.DialURL(config.GetConfig().RedisAddress.String(), redigo.DialKeepAlive(5*time.Minute),
			redigo.DialReadTimeout(5*time.Second), redigo.DialConnectTimeout(5*time.Second))
	},
}

var rdb = redis.NewClient(&redis.Options{
	Addr: config.GetConfig().RedisAddress.Host,
	DB:   0, // use default DB
})

func RedisClient() *redis.Client {
	return rdb
}

func LockClient() *redislock.Client {
	return redislock.New(RedisClient())
}
