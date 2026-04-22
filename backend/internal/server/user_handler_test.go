package server

import (
	"net/http"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_GetProfile_Success(t *testing.T) {
	router := setupTestRouter()
	mockUser := new(mockUserService)
	handler := NewUserHandler(mockUser)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	mockUser.On("GetProfile", uint(1)).Return(&dto.UserResponse{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "pencari",
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/user/profile", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockUser.AssertExpectations(t)
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockUser := new(mockUserService)
	handler := NewUserHandler(mockUser)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	mockUser.On("GetProfile", uint(1)).Return(nil, services.ErrUserNotFound)

	w := makeRequest(t, router, "GET", "/api/v1/user/profile", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "User not found", resp["message"])
	mockUser.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	router := setupTestRouter()
	mockUser := new(mockUserService)
	handler := NewUserHandler(mockUser)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.UpdateProfileRequest{
		Name: strPtr("Updated Name"),
	}

	mockUser.On("UpdateProfile", uint(1), &reqBody).Return(&dto.UserResponse{
		ID:    1,
		Email: "test@example.com",
		Name:  "Updated Name",
		Role:  "pencari",
	}, nil)

	w := makeRequest(t, router, "PUT", "/api/v1/user/profile", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "Profile updated successfully", resp["message"])
	mockUser.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockUser := new(mockUserService)
	handler := NewUserHandler(mockUser)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	w := makeRequest(t, router, "PUT", "/api/v1/user/profile", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp["success"].(bool))
}

func TestUserHandler_UpdateProfile_ServiceError(t *testing.T) {
	router := setupTestRouter()
	mockUser := new(mockUserService)
	handler := NewUserHandler(mockUser)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.UpdateProfileRequest{
		Name: strPtr("Updated Name"),
	}

	mockUser.On("UpdateProfile", uint(1), &reqBody).Return(nil, assert.AnError)

	w := makeRequest(t, router, "PUT", "/api/v1/user/profile", reqBody, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "Failed to update profile", resp["message"])
	mockUser.AssertExpectations(t)
}
