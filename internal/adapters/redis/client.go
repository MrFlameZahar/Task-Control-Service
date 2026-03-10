package redis

import (
	"fmt"

	"TaskControlService/internal/config"

	goredis "github.com/go-redis/redis"
)

func NewClient(cfg *config.Config) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
	})
}