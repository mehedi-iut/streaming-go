package middlewares

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRateLimiter struct {
	Client  *redis.Client
	Window  time.Duration
	Limit   int
	Context context.Context
}

func NewRedisRateLimiter(client *redis.Client, window time.Duration, limit int) *RedisRateLimiter {
	return &RedisRateLimiter{
		Client:  client,
		Window:  window,
		Limit:   limit,
		Context: context.Background(),
	}
}

func (rl *RedisRateLimiter) Allow(userId string) (bool, error) {
	key := "rate_limit:" + userId
	//now := time.Now().Unix()

	pipeline := rl.Client.TxPipeline()

	// Increment the request count
	incr := pipeline.Incr(rl.Context, key)

	// Set the expiration for the key if it's a new key
	pipeline.Expire(rl.Context, key, rl.Window)

	_, err := pipeline.Exec(rl.Context)
	if err != nil {
		return false, err
	}

	// Check if the request count exceeds the limit
	if incr.Val() > int64(rl.Limit) {
		return false, nil
	}
	return true, nil
}
