package server

import (
	"errors"
	"net/http"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Routes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	auth.POST("/register", h.register)
	auth.POST("/login", h.login)
	auth.POST("/refresh", h.refreshToken)
	auth.POST("/logout", h.logout)
	auth.GET("/google", h.googleLogin)
	auth.GET("/google/callback", h.googleCallback)
}

func (h *AuthHandler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			utils.BadRequestResponse(c, "email already registered", nil)
		default:
			utils.InternalServerErrorResponse(c, "something went wrong", err)
		}
		return
	}

	utils.CreatedResponse(c, "user created", response)
}

func (h *AuthHandler) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound),
			errors.Is(err, services.ErrInvalidPassword):
			utils.UnauthorizedResponse(c, "invalid email or password", nil)
		default:
			utils.InternalServerErrorResponse(c, "something went wrong", err)
		}
		return
	}

	utils.SuccessResponse(c, "user logged in", response)
}

func (h *AuthHandler) refreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	response, err := h.authService.RefreshToken(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRefreshToken),
			errors.Is(err, services.ErrRefreshTokenExpired),
			errors.Is(err, services.ErrRefreshTokenRevoked):
			utils.UnauthorizedResponse(c, err.Error(), nil)
		default:
			utils.InternalServerErrorResponse(c, "something went wrong", err)
		}
		return
	}

	utils.SuccessResponse(c, "token refreshed", response)
}

func (h *AuthHandler) logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		utils.InternalServerErrorResponse(c, "something went wrong", err)
		return
	}

	utils.SuccessResponse(c, "user logged out", nil)
}

func (h *AuthHandler) googleLogin(c *gin.Context) {
	url := h.authService.GoogleLogin()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) googleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.BadRequestResponse(c, "missing oauth code", nil)
		return
	}

	response, err := h.authService.GoogleCallback(code)
	if err != nil {
		utils.InternalServerErrorResponse(c, "google oauth failed", err)
		return
	}

	utils.SuccessResponse(c, "google login successful", response)
}
