package utils

import (
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTokenPair_Success(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key",
		ExpiresIn:        time.Hour,
		RefreshExpiresIn: time.Hour * 24,
	}

	accessToken, refreshToken, err := GenerateTokenPair(cfg, 1, "test@example.com", "pencari")

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestGenerateTokenPair_TokensAreValidJWT(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key",
		ExpiresIn:        time.Hour,
		RefreshExpiresIn: time.Hour * 24,
	}

	accessToken, refreshToken, err := GenerateTokenPair(cfg, 1, "test@example.com", "pencari")
	require.NoError(t, err)

	_, err = jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	assert.NoError(t, err)

	_, err = jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	assert.NoError(t, err)
}

func TestValidateToken_Success(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key",
		ExpiresIn:        time.Hour,
		RefreshExpiresIn: time.Hour * 24,
	}

	accessToken, _, err := GenerateTokenPair(cfg, 1, "test@example.com", "pencari")
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken, cfg.Secret)

	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "pencari", claims.Role)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key",
		ExpiresIn:        time.Hour,
		RefreshExpiresIn: time.Hour * 24,
	}

	accessToken, _, err := GenerateTokenPair(cfg, 1, "test@example.com", "pencari")
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken, "wrong-secret")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key",
		ExpiresIn:        -time.Hour,
		RefreshExpiresIn: -time.Hour * 24,
	}

	accessToken, _, err := GenerateTokenPair(cfg, 1, "test@example.com", "pencari")
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken, cfg.Secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	claims, err := ValidateToken("not-a-valid-jwt", "secret")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	claims, err := ValidateToken("", "secret")

	assert.Error(t, err)
	assert.Nil(t, claims)
}
