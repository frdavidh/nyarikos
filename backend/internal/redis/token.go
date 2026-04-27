package redis

import (
	"context"
	"fmt"
	"time"
)

type TokenStore struct {
	client *Client
}

func NewTokenStore(client *Client) *TokenStore {
	return &TokenStore{client: client}
}

func (t *TokenStore) SaveRefreshToken(ctx context.Context, userID uint, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	value := fmt.Sprintf("%d", userID)

	return t.client.rdb.Set(ctx, key, value, ttl).Err()
}

func (t *TokenStore) GetUserIDByRefreshToken(ctx context.Context, token string) (uint, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	val, err := t.client.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("refresh token not found: %w", err)
	}

	var userID uint
	fmt.Sscanf(val, "%d", &userID)
	return userID, nil
}

func (t *TokenStore) RevokeRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return t.client.rdb.Del(ctx, key).Err()
}

// func (t *TokenStore) RevokeAllUserTokens(ctx context.Context, userID uint) error {
// 	retun nil
// }
