package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type mockHTTPClient struct {
	resp *http.Response
	err  error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func newMockHTTPClient(body string, statusCode int) *mockHTTPClient {
	return &mockHTTPClient{
		resp: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		},
	}
}

type mockOAuthExchanger struct {
	token *oauth2.Token
	err   error
}

func (m *mockOAuthExchanger) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.token, m.err
}

func newAuthServiceWithMockHTTP(db *gorm.DB, cfg *config.Config, client httpClient, exchanger oauthExchanger) *authService {
	return &authService{
		db:            db,
		config:        cfg,
		httpClient:    client,
		oauthExchange: exchanger,
	}
}

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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
	service := NewAuthService(db, cfg, nil, nil)

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

	req := &dto.RefreshTokenRequest{RefreshToken: refresh}
	resp, err := service.RefreshToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.Email, resp.User.Email)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg, nil, nil)

	req := &dto.RefreshTokenRequest{RefreshToken: "invalid-token"}
	resp, err := service.RefreshToken(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidRefreshToken, err)
}

func TestAuthService_Logout(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg, nil, nil)

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

	err = service.Logout(context.Background(), refresh)

	assert.NoError(t, err)
}

func TestAuthService_GoogleLogin(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()
	service := NewAuthService(db, cfg, nil, nil)

	url, err := service.GoogleLogin(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "accounts.google.com")
}

func TestAuthService_GoogleCallback_ExistingSocialAccount(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	user := models.User{
		Email:    "google@example.com",
		Name:     "Google User",
		Role:     models.RolePencari,
		IsActive: true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	social := models.SocialAccount{
		UserID:       user.ID,
		ProviderName: models.ProviderGoogle,
		ProviderID:   "google-123",
	}
	err = db.Create(&social).Error
	require.NoError(t, err)

	userInfo := map[string]string{
		"id":    "google-123",
		"email": "google@example.com",
		"name":  "Google User",
	}
	body, _ := json.Marshal(userInfo)
	mockClient := newMockHTTPClient(string(body), http.StatusOK)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}

	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "google@example.com", resp.User.Email)
}

func TestAuthService_GoogleCallback_ExistingUserLinkSocial(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	user := models.User{
		Email:    "existing@example.com",
		Name:     "Existing User",
		Role:     models.RolePencari,
		IsActive: true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	userInfo := map[string]string{
		"id":    "google-456",
		"email": "existing@example.com",
		"name":  "Existing User",
	}
	body, _ := json.Marshal(userInfo)
	mockClient := newMockHTTPClient(string(body), http.StatusOK)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}

	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "existing@example.com", resp.User.Email)

	var social models.SocialAccount
	err = db.Where("provider_id = ?", "google-456").First(&social).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, social.UserID)
}

func TestAuthService_GoogleCallback_NewUser(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	userInfo := map[string]string{
		"id":    "google-789",
		"email": "newuser@example.com",
		"name":  "New User",
	}
	body, _ := json.Marshal(userInfo)
	mockClient := newMockHTTPClient(string(body), http.StatusOK)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}

	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "newuser@example.com", resp.User.Email)
	assert.Equal(t, "New User", resp.User.Name)
	assert.True(t, resp.User.IsActive)

	var user models.User
	err = db.Where("email = ?", "newuser@example.com").First(&user).Error
	require.NoError(t, err)
	assert.Equal(t, "New User", user.Name)

	var social models.SocialAccount
	err = db.Where("provider_id = ?", "google-789").First(&social).Error
	require.NoError(t, err)
	assert.Equal(t, user.ID, social.UserID)
}

func TestAuthService_GoogleCallback_IncompleteUserInfo(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	userInfo := map[string]string{
		"id":    "",
		"email": "",
		"name":  "Incomplete",
	}
	body, _ := json.Marshal(userInfo)
	mockClient := newMockHTTPClient(string(body), http.StatusOK)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}

	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "incomplete user info")
}

func TestAuthService_GoogleCallback_HTTPError(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	mockClient := &mockHTTPClient{resp: nil, err: errors.New("network error")}
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}
	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to fetch google user info")
}

func TestAuthService_GoogleCallback_Non200Status(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	mockClient := newMockHTTPClient("error", http.StatusUnauthorized)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}
	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "google userinfo returned status")
}

func TestAuthService_GoogleCallback_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	cfg := testConfig()

	mockClient := newMockHTTPClient("not-json", http.StatusOK)
	mockOAuth := &mockOAuthExchanger{token: &oauth2.Token{AccessToken: "access-token"}, err: nil}
	service := newAuthServiceWithMockHTTP(db, cfg, mockClient, mockOAuth)

	resp, err := service.GoogleCallback(context.Background(), "auth-code", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestIsDuplicateKeyError_Nil(t *testing.T) {
	assert.False(t, isDuplicateKeyError(nil))
}

func TestIsDuplicateKeyError_GormDuplicatedKey(t *testing.T) {
	assert.True(t, isDuplicateKeyError(gorm.ErrDuplicatedKey))
}

func TestIsDuplicateKeyError_UniqueConstraint(t *testing.T) {
	assert.True(t, isDuplicateKeyError(errors.New("UNIQUE constraint failed: users.email")))
}

func TestIsDuplicateKeyError_DuplicateKeyValue(t *testing.T) {
	assert.True(t, isDuplicateKeyError(errors.New("duplicate key value violates unique constraint")))
}

func TestIsDuplicateKeyError_OtherError(t *testing.T) {
	assert.False(t, isDuplicateKeyError(errors.New("some other error")))
}

func strPtr(s string) *string {
	return &s
}
