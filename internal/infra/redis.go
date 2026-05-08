package infra

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(conf config.RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
}

func CheckRedis(client *redis.Client) DependencyCheck {
	return DependencyCheck{
		Name: "redis",
		Check: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := client.Ping(pingCtx).Err(); err != nil {
				return fmt.Errorf("ping redis: %w", err)
			}

			return nil
		},
	}
}
