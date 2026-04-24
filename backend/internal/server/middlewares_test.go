package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupMiddlewareServer(cfg *config.Config) *Server {
	return New(cfg, nil, nil, nil, nil, nil, nil, nil, nil)
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpiresIn: time.Hour},
	}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(srv.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpiresIn: time.Hour},
	}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(srv.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic token123")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpiresIn: time.Hour},
	}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(srv.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpiresIn: time.Hour},
	}
	srv := setupMiddlewareServer(cfg)

	token, _, err := utils.GenerateTokenPair(&cfg.JWT, 42, "test@example.com", "pencari")
	assert.NoError(t, err)

	var capturedUserID uint
	var capturedEmail, capturedRole string

	router := gin.New()
	router.Use(srv.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		capturedUserID = c.GetUint("user_id")
		capturedEmail = c.GetString("email")
		capturedRole = c.GetString("role")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(42), capturedUserID)
	assert.Equal(t, "test@example.com", capturedEmail)
	assert.Equal(t, "pencari", capturedRole)
}

func TestAuthMiddleware_MalformedBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpiresIn: time.Hour},
	}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(srv.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer") // missing token part
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRoleMiddleware_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "pemilik")
		c.Next()
	})
	router.Use(srv.roleMiddleware("pemilik"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleMiddleware_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	srv := setupMiddlewareServer(cfg)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "pencari")
		c.Next()
	})
	router.Use(srv.roleMiddleware("pemilik"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
