package database

import (
	"context"
	"drivo/internal/config"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func Redis(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	// cfg, err := config.Load()

	opt, err := redis.ParseURL(cfg.REDIS_URL)

	if err != nil {
		return nil, fmt.Errorf("Error parsing url to redis: %v", err)

	}

	redisClient := redis.NewClient(opt)

	// Verify connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Redis: %v", err)

	}

	fmt.Println("Connected to Upstash Redis!")

	return redisClient, nil

}
