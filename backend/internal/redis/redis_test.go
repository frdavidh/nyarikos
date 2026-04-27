package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Connection(t *testing.T) {
	client, err := New("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available (is docker-compose up -d redis running?): %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("ping", func(t *testing.T) {
		err := client.RDB().Ping(ctx).Err()
		assert.NoError(t, err)
	})

	t.Run("set and get", func(t *testing.T) {
		key := "test:set_get"
		err := client.RDB().Set(ctx, key, "hello", 5*time.Second).Err()
		require.NoError(t, err)

		val, err := client.RDB().Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, "hello", val)
	})

	t.Run("ttl expiration", func(t *testing.T) {
		key := "test:expire"
		err := client.RDB().Set(ctx, key, "value", 1*time.Second).Err()
		require.NoError(t, err)

		_, err = client.RDB().Get(ctx, key).Result()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		_, err = client.RDB().Get(ctx, key).Result()
		assert.Error(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		key := "test:delete"
		err := client.RDB().Set(ctx, key, "value", 10*time.Second).Err()
		require.NoError(t, err)

		n, err := client.RDB().Del(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)

		_, err = client.RDB().Get(ctx, key).Result()
		assert.Error(t, err)
	})
}

func TestLock(t *testing.T) {
	client, err := New("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("acquire and release", func(t *testing.T) {
		key := "test:lock:1"
		unlock, err := client.Lock(ctx, key, 5*time.Second)
		require.NoError(t, err)
		assert.NotNil(t, unlock)

		val, err := client.RDB().Get(ctx, key).Result()
		require.NoError(t, err)
		assert.NotEmpty(t, val)

		err = unlock()
		assert.NoError(t, err)

		_, err = client.RDB().Get(ctx, key).Result()
		assert.Error(t, err)
	})

	t.Run("double lock fails", func(t *testing.T) {
		key := "test:lock:2"
		unlock, err := client.Lock(ctx, key, 5*time.Second)
		require.NoError(t, err)
		defer unlock()

		_, err = client.Lock(ctx, key, 5*time.Second)
		assert.Equal(t, ErrLockNotAcquired, err)
	})

	t.Run("lock expires automatically", func(t *testing.T) {
		key := "test:lock:3"
		unlock, err := client.Lock(ctx, key, 500*time.Millisecond)
		require.NoError(t, err)
		defer unlock()

		time.Sleep(1 * time.Second)

		unlock2, err := client.Lock(ctx, key, 5*time.Second)
		require.NoError(t, err)
		unlock2()
	})
}

func TestTokenStore(t *testing.T) {
	client, err := New("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	store := NewTokenStore(client)

	t.Run("save and get", func(t *testing.T) {
		token := "test-token-abc123"
		err := store.SaveRefreshToken(ctx, 42, token, 5*time.Second)
		require.NoError(t, err)

		userID, err := store.GetUserIDByRefreshToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, uint(42), userID)
	})

	t.Run("get missing token", func(t *testing.T) {
		_, err := store.GetUserIDByRefreshToken(ctx, "nonexistent-token")
		assert.Error(t, err)
	})

	t.Run("revoke token", func(t *testing.T) {
		token := "test-token-to-revoke"
		err := store.SaveRefreshToken(ctx, 99, token, 5*time.Minute)
		require.NoError(t, err)

		err = store.RevokeRefreshToken(ctx, token)
		require.NoError(t, err)

		_, err = store.GetUserIDByRefreshToken(ctx, token)
		assert.Error(t, err)
	})

	t.Run("token auto expires", func(t *testing.T) {
		token := "test-token-expire"
		err := store.SaveRefreshToken(ctx, 7, token, 1*time.Second)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		_, err = store.GetUserIDByRefreshToken(ctx, token)
		assert.Error(t, err)
	})
}
