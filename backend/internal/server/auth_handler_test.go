package server

import (
	"errors"
	"net/http"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockAuth.On("Register", &reqBody).Return(&dto.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User: dto.UserResponse{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  "pencari",
		},
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/auth/register", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "user created", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	w := makeRequest(t, router, "POST", "/api/v1/auth/register", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp["success"].(bool))
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockAuth.On("Register", &reqBody).Return(nil, services.ErrEmailAlreadyExists)

	w := makeRequest(t, router, "POST", "/api/v1/auth/register", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "email already registered", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockAuth.On("Login", &reqBody).Return(&dto.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User: dto.UserResponse{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		},
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/auth/login", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "user logged in", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockAuth.On("Login", &reqBody).Return(nil, services.ErrInvalidPassword)

	w := makeRequest(t, router, "POST", "/api/v1/auth/login", reqBody, nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "invalid email or password", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.RefreshTokenRequest{RefreshToken: "old-refresh-token"}

	mockAuth.On("RefreshToken", &reqBody).Return(&dto.AuthResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/auth/refresh", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "token refreshed", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.RefreshTokenRequest{RefreshToken: "invalid-token"}

	mockAuth.On("RefreshToken", &reqBody).Return(nil, services.ErrInvalidRefreshToken)

	w := makeRequest(t, router, "POST", "/api/v1/auth/refresh", reqBody, nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, services.ErrInvalidRefreshToken.Error(), resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	reqBody := dto.RefreshTokenRequest{RefreshToken: "refresh-token"}

	mockAuth.On("Logout", "refresh-token").Return(nil)

	w := makeRequest(t, router, "POST", "/api/v1/auth/logout", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "user logged out", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_GoogleLogin(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	mockAuth.On("GoogleLogin").Return("https://accounts.google.com/oauth/authorize?...")

	w := makeRequest(t, router, "GET", "/api/v1/auth/google", nil, nil)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "accounts.google.com")
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_GoogleCallback_Success(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	mockAuth.On("GoogleCallback", "auth-code").Return(&dto.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User: dto.UserResponse{
			ID:    1,
			Email: "test@gmail.com",
			Name:  "Test User",
		},
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/auth/google/callback?code=auth-code", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "google login successful", resp["message"])
	mockAuth.AssertExpectations(t)
}

func TestAuthHandler_GoogleCallback_MissingCode(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	w := makeRequest(t, router, "GET", "/api/v1/auth/google/callback", nil, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "missing oauth code", resp["message"])
}

func TestAuthHandler_GoogleCallback_ServiceError(t *testing.T) {
	router := setupTestRouter()
	mockAuth := new(mockAuthService)
	handler := NewAuthHandler(mockAuth)

	api := router.Group("/api/v1")
	handler.Routes(api)

	mockAuth.On("GoogleCallback", "auth-code").Return(nil, errors.New("oauth failed"))

	w := makeRequest(t, router, "GET", "/api/v1/auth/google/callback?code=auth-code", nil, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "google oauth failed", resp["message"])
	mockAuth.AssertExpectations(t)
}
