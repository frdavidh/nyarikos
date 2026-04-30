package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = fmt.Errorf("failed to acquire lock")

func (c *Client) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	err := c.rdb.SetArgs(ctx, key, token, goredis.SetArgs{Mode: "NX", TTL: ttl}).Err()
	if err == goredis.Nil {
		return nil, ErrLockNotAcquired
	}
	if err != nil {
		return nil, fmt.Errorf("lock: %w", err)
	}

	release := func() error {
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		_, err := c.rdb.Eval(context.Background(), script, []string{key}, token).Result()
		if err != nil {
			return fmt.Errorf("unlock: %w", err)
		}
		return nil
	}

	return release, nil
}
