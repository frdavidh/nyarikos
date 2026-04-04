package server

import (
	"net/http"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

func (s *Server) register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.Register(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	utils.CreatedResponse(c, "user created", response)
}

func (s *Server) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.Login(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	utils.SuccessResponse(c, "user logged in", response)
}

func (s *Server) refreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.RefreshToken(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	utils.SuccessResponse(c, "token refreshed", response)
}

func (s *Server) logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	if err := authService.Logout(req.RefreshToken); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
		return
	}

	utils.SuccessResponse(c, "user logged out", nil)
}
