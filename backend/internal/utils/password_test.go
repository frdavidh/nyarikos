package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestHashPassword_Success(t *testing.T) {
	hash, err := HashPassword("password123", DefaultParams)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("password123", DefaultParams)
	require.NoError(t, err)

	match, err := VerifyPassword("password123", hash)

	require.NoError(t, err)
	assert.True(t, match)
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("password123", DefaultParams)
	require.NoError(t, err)

	match, err := VerifyPassword("wrongpassword", hash)

	require.NoError(t, err)
	assert.False(t, match)
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	tests := []struct {
		name  string
		hash  string
		match bool
	}{
		{"too few parts", "$argon2id$v=19$m=65536,t=3,p=2$salt", false},
		{"too many parts", "$argon2id$v=19$m=65536,t=3,p=2$salt$hash$extra", false},
		{"wrong algorithm", "$bcrypt$v=19$m=65536,t=3,p=2$salt$hash", false},
		{"invalid version format", "$argon2id$version=19$m=65536,t=3,p=2$salt$hash", false},
		{"mismatched version", "$argon2id$v=999$m=65536,t=3,p=2$c2FsdA$hash", false},
		{"invalid params format", "$argon2id$v=19$badparams$salt$hash", false},
		{"invalid salt base64", "$argon2id$v=19$m=65536,t=3,p=2$!!!badbase64!!!$hash", false},
		{"invalid hash base64", "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$!!!badbase64!!!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := VerifyPassword("password123", tt.hash)
			assert.Error(t, err)
			assert.False(t, match)
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := generateSalt(16)
	require.NoError(t, err)
	assert.Len(t, salt1, 16)

	salt2, err := generateSalt(16)
	require.NoError(t, err)
	assert.Len(t, salt2, 16)
	assert.NotEqual(t, salt1, salt2, "salt should be random")
}

func TestParseHash_Success(t *testing.T) {
	hash, err := HashPassword("password123", DefaultParams)
	require.NoError(t, err)

	params, salt, decodedHash, err := parseHash(hash)

	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.NotEmpty(t, salt)
	assert.NotEmpty(t, decodedHash)
	assert.Equal(t, DefaultParams.Memory, params.Memory)
	assert.Equal(t, DefaultParams.Iterations, params.Iterations)
	assert.Equal(t, DefaultParams.Parallelism, params.Parallelism)
	assert.Equal(t, DefaultParams.KeyLength, params.KeyLength)
}

func TestParseHash_UnsupportedVersion(t *testing.T) {
	salt := "c2FsdA"
	hash := "aGFzaA"
	badHash := "$argon2id$v=1$m=65536,t=3,p=2$" + salt + "$" + hash

	_, _, _, err := parseHash(badHash)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version")
}

func TestHashPasswordAndVerifyRoundTrip(t *testing.T) {
	passwords := []string{
		"short",
		"a-very-long-password-with-many-characters-and-symbols!@#$%",
		"password with spaces",
		"1234567890",
	}

	for _, pwd := range passwords {
		t.Run("password_"+pwd[:min(len(pwd), 10)], func(t *testing.T) {
			hash, err := HashPassword(pwd, DefaultParams)
			require.NoError(t, err)

			match, err := VerifyPassword(pwd, hash)
			require.NoError(t, err)
			assert.True(t, match)

			match, err = VerifyPassword(pwd+"wrong", hash)
			require.NoError(t, err)
			assert.False(t, match)
		})
	}
}

func TestHashPassword_UsesArgon2id(t *testing.T) {
	hash, err := HashPassword("password", DefaultParams)
	require.NoError(t, err)
	assert.Contains(t, hash, "argon2id")
}

func TestVerifyPassword_InvalidVersion(t *testing.T) {
	invalidHash := "$argon2id$v=" + string(rune('0'+argon2.Version+1)) + "$m=65536,t=3,p=2$c2FsdA$hash"
	match, err := VerifyPassword("password", invalidHash)
	assert.Error(t, err)
	assert.False(t, match)
}
