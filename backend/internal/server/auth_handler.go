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

// @Tags			Auth
// @Summary		Register a new user
// @Description	Register a new user with email, password, and optional details
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RegisterRequest						true	"Register Request"
// @Success		201		{object}	utils.Response{data=dto.AuthResponse}	"User created"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/auth/register [post]
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

// @Tags			Auth
// @Summary		Login user
// @Description	Login with email and password to get access and refresh tokens
// @Accept			json
// @Produce		json
// @Param			request	body		dto.LoginRequest						true	"Login Request"
// @Success		200		{object}	utils.Response{data=dto.AuthResponse}	"User logged in"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Invalid credentials"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/auth/login [post]
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

// @Tags			Auth
// @Summary		Refresh access token
// @Description	Get a new access token using a refresh token
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RefreshTokenRequest					true	"Refresh Token Request"
// @Success		200		{object}	utils.Response{data=dto.AuthResponse}	"Token refreshed"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Invalid or expired refresh token"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/auth/refresh [post]
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

// @Tags			Auth
// @Summary		Logout user
// @Description	Logout user and revoke refresh token
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RefreshTokenRequest	true	"Logout Request"
// @Success		200		{object}	utils.Response			"User logged out"
// @Failure		400		{object}	utils.Response			"Invalid request"
// @Failure		500		{object}	utils.Response			"Internal server error"
// @Router			/auth/logout [post]
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

// @Tags			Auth
// @Summary		Google OAuth login
// @Description	Redirect to Google OAuth login page
// @Produce		json
// @Success		307	"Redirect to Google"
// @Router			/auth/google [get]
func (h *AuthHandler) googleLogin(c *gin.Context) {
	url := h.authService.GoogleLogin()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// @Tags			Auth
// @Summary		Google OAuth callback
// @Description	Handle Google OAuth callback and authenticate user
// @Produce		json
// @Param			code	query		string									true	"OAuth authorization code"
// @Success		200		{object}	utils.Response{data=dto.AuthResponse}	"Google login successful"
// @Failure		400		{object}	utils.Response							"Missing OAuth code"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/auth/google/callback [get]
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
