package redis

import (
	"context"
	"fmt"
	"time"
)

var ErrLockNotAcquired = fmt.Errorf("failed to acquire lock")

func (c *Client) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	ok, err := c.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("lock: %w", err)
	}
	if !ok {
		return nil, ErrLockNotAcquired
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
