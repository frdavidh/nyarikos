package services

import (
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"gorm.io/gorm"
)

type UserService interface {
	GetProfile(userId uint) (*dto.UserResponse, error)
	UpdateProfile(userId uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type userService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

func (s *userService) GetProfile(userId uint) (*dto.UserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (s *userService) UpdateProfile(userId uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = req.PhoneNumber
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return s.GetProfile(userId)
}
