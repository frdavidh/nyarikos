package services

import (
	"errors"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthService(db *gorm.DB, config *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: config,
	}
}

func (s *AuthService) generateAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := utils.GenerateTokenPair(
		&s.config.JWT,
		user.ID,
		user.Email,
		string(user.Role),
	)
	if err != nil {
		return nil, err
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
		},
	}, nil
}

func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already registered")
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

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	match, err := utils.VerifyPassword(req.Password, *user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, errors.New("invalid password")
	}

	return s.generateAuthResponse(&user)
}

func (s *AuthService) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken, []byte(s.config.JWT.Secret))
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var refreshToken models.RefreshToken
	if err := s.db.Where("token = ?", req.RefreshToken).First(&refreshToken).Error; err != nil {
		return nil, errors.New("refresh token not found")
	}

	if refreshToken.IsRevoked != nil && *refreshToken.IsRevoked {
		return nil, errors.New("refresh token is revoked")
	}

	if refreshToken.ExpiredAt.Before(time.Now()) {
		return nil, errors.New("refresh token is expired")
	}

	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return s.generateAuthResponse(&user)
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.db.Model(&models.RefreshToken{}).Where("token = ?", refreshToken).Update("is_revoked", true).Error
}
