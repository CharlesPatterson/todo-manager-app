package model

import (
	"fmt"
	"os"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	Store            *persist.RedisStore
	DefaultCacheTime time.Duration
}

func SetupRedisCache() *RedisCache {
	return &RedisCache{
		Store: persist.NewRedisStore(redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr: fmt.Sprintf(
				"%s:%s",
				os.Getenv("REDIS_HOST"),
				os.Getenv("REDIS_PORT"),
			),
		})),
		DefaultCacheTime: 15 * time.Minute,
	}
}
