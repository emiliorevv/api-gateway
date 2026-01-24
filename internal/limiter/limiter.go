package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local last_tokens = tonumber(redis.call("HGET", key, "tokens"))
local last_refill = tonumber(redis.call("HGET", key, "last_refill"))

if last_tokens == nil then
last_tokens = capacity
last_refill = now
end

local delta = math.max(0, now - last_refill)
local filled_tokens = math.min(capacity, last_tokens + (delta * rate))

if filled_tokens < requested then
return 0 -- False (Blocked)
end

redis.call("HSET", key, "tokens", filled_tokens - requested, "last_refill", now)
redis.call("EXPIRE", key, 60)
return 1 -- True (Accepted)
`

type RateLimiter struct {
	client   *redis.Client


}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client:   client,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, rate float64 ) (bool, error) {
	redisKey := fmt.Sprintf("rate_limit_%s", key)

	now := time.Now().Unix()

	result, err := rl.client.Eval(ctx, tokenBucketScript, []string{redisKey}, limit, rate, now, 1).Result()

	if err != nil {
		return false, fmt.Errorf("error executing lua script: %w", err)
	}

	return result.(int64) == 1, nil
}

func NewRedisClient(addr string) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()

		return nil, fmt.Errorf("redis mock connection error: %w", err)

	}
	return rdb, nil
}
