package services

import (
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestRefreshToken(cfg *config.JWTConfig, userID uint, email, role string) (string, error) {
	refreshClaims := &utils.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Second)),
		},
	}
	rf := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	return rf.SignedString([]byte(cfg.Secret))
}

func TestAuthService_Register_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	req := &dto.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		Name:        "Test User",
		PhoneNumber: "08123456789",
		Role:        "pencari",
	}

	resp, err := service.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.Equal(t, "pencari", resp.User.Role)
	assert.True(t, resp.User.IsActive)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	existing := models.User{
		Email:    "test@example.com",
		Name:     "Existing",
		Password: strPtr("hashed"),
		Role:     models.RolePencari,
	}
	err := db.Create(&existing).Error
	require.NoError(t, err)

	req := &dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	resp, err := service.Register(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrEmailAlreadyExists, err)
}

func TestAuthService_Register_PemilikRole(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	req := &dto.RegisterRequest{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner",
		Role:     "pemilik",
	}

	resp, err := service.Register(req)

	assert.NoError(t, err)
	assert.Equal(t, "pemilik", resp.User.Role)
}

func TestAuthService_Login_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	hashed, err := utils.HashPassword("password123", utils.DefaultParams)
	require.NoError(t, err)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: &hashed,
		Role:     models.RolePencari,
		IsActive: true,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := service.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.Email, resp.User.Email)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	req := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	resp, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	hashed, err := utils.HashPassword("correctpassword", utils.DefaultParams)
	require.NoError(t, err)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: &hashed,
		Role:     models.RolePencari,
		IsActive: true,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	resp, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidPassword, err)
}

func TestAuthService_Login_UserInactive(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	hashed, err := utils.HashPassword("password123", utils.DefaultParams)
	require.NoError(t, err)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: &hashed,
		Role:     models.RolePencari,
		IsActive: true,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	err = db.Model(&user).Update("is_active", false).Error
	require.NoError(t, err)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUserInactive, err)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     models.RolePencari,
		IsActive: true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	refresh, err := generateTestRefreshToken(&cfg.JWT, user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(cfg.JWT.RefreshExpiresIn),
	}
	err = db.Create(&rt).Error
	require.NoError(t, err)

	req := &dto.RefreshTokenRequest{RefreshToken: refresh}
	resp, err := service.RefreshToken(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.Email, resp.User.Email)

	var updatedRT models.RefreshToken
	err = db.First(&updatedRT, rt.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updatedRT.IsRevoked)
	assert.True(t, *updatedRT.IsRevoked)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	req := &dto.RefreshTokenRequest{RefreshToken: "invalid-token"}
	resp, err := service.RefreshToken(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidRefreshToken, err)
}

func TestAuthService_RefreshToken_Revoked(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     models.RolePencari,
		IsActive: true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	refresh, err := generateTestRefreshToken(&cfg.JWT, user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	revoked := true
	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(cfg.JWT.RefreshExpiresIn),
		IsRevoked: &revoked,
	}
	err = db.Create(&rt).Error
	require.NoError(t, err)

	req := &dto.RefreshTokenRequest{RefreshToken: refresh}
	resp, err := service.RefreshToken(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrRefreshTokenRevoked, err)
}

func TestAuthService_Logout(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	user := models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     models.RolePencari,
		IsActive: true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	_, refresh, err := utils.GenerateTokenPair(
		&cfg.JWT, user.ID, user.Email, string(user.Role),
	)
	require.NoError(t, err)

	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(cfg.JWT.RefreshExpiresIn),
	}
	err = db.Create(&rt).Error
	require.NoError(t, err)

	err = service.Logout(refresh)

	assert.NoError(t, err)

	var updated models.RefreshToken
	err = db.First(&updated, rt.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updated.IsRevoked)
	assert.True(t, *updated.IsRevoked)
}

func TestAuthService_GoogleLogin(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg)

	url := service.GoogleLogin()

	assert.NotEmpty(t, url)
	assert.Contains(t, url, "accounts.google.com")
}

func strPtr(s string) *string {
	return &s
}
