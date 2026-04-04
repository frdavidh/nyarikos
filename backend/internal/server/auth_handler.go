package server

import (
	"errors"

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

	response, err := s.authService.Register(&req)
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

func (s *Server) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	response, err := s.authService.Login(&req)
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

func (s *Server) refreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	response, err := s.authService.RefreshToken(&req)
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

func (s *Server) logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	if err := s.authService.Logout(req.RefreshToken); err != nil {
		utils.InternalServerErrorResponse(c, "something went wrong", err)
		return
	}

	utils.SuccessResponse(c, "user logged out", nil)
}
