package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenRevoked = errors.New("refresh token is revoked")
	ErrRefreshTokenExpired = errors.New("refresh token is expired")
	ErrUserInactive        = errors.New("user is inactive")
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(refreshToken string) error
}

type authService struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthService(db *gorm.DB, config *config.Config) AuthService {
	return &authService{
		db:     db,
		config: config,
	}
}

func (s *authService) generateAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := utils.GenerateTokenPair(
		&s.config.JWT,
		user.ID,
		user.Email,
		string(user.Role),
	)
	if err != nil {
		return nil, err
	}

	refreshTokenRecord := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.config.JWT.RefreshExpiresIn),
	}
	if err := s.db.Create(&refreshTokenRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password, utils.DefaultParams)
	if err != nil {
		return nil, err
	}

	role := models.RolePencari
	if req.Role == string(models.RolePemilik) {
		role = models.RolePemilik
	}

	user := models.User{
		Email:       req.Email,
		Password:    &hashedPassword,
		Name:        req.Name,
		PhoneNumber: &req.PhoneNumber,
		Role:        role,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	match, err := utils.VerifyPassword(req.Password, *user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidPassword
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken, []byte(s.config.JWT.Secret))
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	var refreshToken models.RefreshToken
	if err := s.db.Where("token = ?", req.RefreshToken).First(&refreshToken).Error; err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if refreshToken.IsRevoked != nil && *refreshToken.IsRevoked {
		return nil, ErrRefreshTokenRevoked
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.db.Model(&refreshToken).Update("is_revoked", true).Error; err != nil {
		return nil, err
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) Logout(refreshToken string) error {
	return s.db.Model(&models.RefreshToken{}).Where("token = ?", refreshToken).Update("is_revoked", true).Error
}
