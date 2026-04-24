package services

import (
	"errors"
	"fmt"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrKostNotFound    = errors.New("record not found")
	ErrInvalidFileType = errors.New("invalid file type")
)

type KostService interface {
	CreateKost(ownerID uint, req *dto.CreateKostRequest) (*dto.KostResponse, error)
	UpdateKost(kostID, userID uint, req *dto.UpdateKostRequest) (*dto.KostResponse, error)
	DeleteKost(kostID, userID uint) (*dto.KostResponse, error)
	GetAllKost(page, limit int) ([]dto.KostResponse, int64, error)
	GetKost(id uint) (*dto.KostResponse, error)
	AddKostImage(kostID, userID uint, url, altText string) error
}

type kostService struct {
	db *gorm.DB
}

func NewKostService(db *gorm.DB) KostService {
	return &kostService{db: db}
}

func (s *kostService) CreateKost(ownerID uint, req *dto.CreateKostRequest) (*dto.KostResponse, error) {
	isPremium := false
	if req.IsPremium != nil {
		isPremium = *req.IsPremium
	}

	kost := models.Kost{
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: &req.Description,
		Address:     req.Address,
		City:        req.City,
		IsPremium:   isPremium,
		KostType:    req.KostType,
	}

	if err := s.db.Create(&kost).Error; err != nil {
		return nil, fmt.Errorf("failed to create kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

func (s *kostService) GetAllKost(page, limit int) ([]dto.KostResponse, int64, error) {
	var kosts []models.Kost
	var total int64

	if err := s.db.Model(&models.Kost{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count kost: %w", err)
	}

	offset := (page - 1) * limit
	if err := s.db.Preload("Images").Offset(offset).Limit(limit).Find(&kosts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get all kost: %w", err)
	}

	responses := make([]dto.KostResponse, 0, len(kosts))
	for i := range kosts {
		responses = append(responses, *toKostResponse(&kosts[i]))
	}

	return responses, total, nil
}

func (s *kostService) GetKost(id uint) (*dto.KostResponse, error) {
	var kost models.Kost
	if err := s.db.Preload("Images").First(&kost, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrKostNotFound
		}
		return nil, fmt.Errorf("failed to get kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

func (s *kostService) UpdateKost(kostID, userID uint, req *dto.UpdateKostRequest) (*dto.KostResponse, error) {
	var kost models.Kost
	if err := s.db.First(&kost, kostID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrKostNotFound
		}
		return nil, fmt.Errorf("failed to get kost: %w", err)
	}

	if kost.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	if req.Name != nil {
		kost.Name = *req.Name
	}
	if req.Description != nil {
		kost.Description = req.Description
	}
	if req.Address != nil {
		kost.Address = *req.Address
	}
	if req.City != nil {
		kost.City = *req.City
	}
	if req.IsPremium != nil {
		kost.IsPremium = *req.IsPremium
	}
	if req.KostType != nil {
		kost.KostType = *req.KostType
	}

	if err := s.db.Save(&kost).Error; err != nil {
		return nil, fmt.Errorf("failed to update kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

func (s *kostService) DeleteKost(kostID, userID uint) (*dto.KostResponse, error) {
	var kost models.Kost
	if err := s.db.First(&kost, kostID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrKostNotFound
		}
		return nil, fmt.Errorf("failed to get kost: %w", err)
	}

	if kost.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	if err := s.db.Delete(&kost).Error; err != nil {
		return nil, fmt.Errorf("failed to delete kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

func (s *kostService) AddKostImage(kostID, userID uint, url, altText string) error {
	var kost models.Kost
	if err := s.db.First(&kost, kostID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrKostNotFound
		}
		return fmt.Errorf("failed to get kost: %w", err)
	}

	if kost.OwnerID != userID {
		return ErrUnauthorized
	}

	var count int64
	if err := s.db.Model(&models.KostImage{}).Where("kost_id = ?", kostID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count images: %w", err)
	}

	image := models.KostImage{
		KostID:   kostID,
		ImageURL: url,
		AltText:  altText,
		IsMain:   count == 0,
	}

	return s.db.Create(&image).Error
}

func toKostResponse(kost *models.Kost) *dto.KostResponse {
	images := make([]dto.ImageResponse, 0, len(kost.Images))
	for _, img := range kost.Images {
		images = append(images, dto.ImageResponse{
			ID:       img.ID,
			ImageURL: img.ImageURL,
		})
	}
	return &dto.KostResponse{
		ID:          kost.ID,
		OwnerID:     kost.OwnerID,
		Name:        kost.Name,
		Description: kost.Description,
		Address:     kost.Address,
		City:        kost.City,
		IsPremium:   kost.IsPremium,
		KostType:    kost.KostType,
		CreatedAt:   kost.CreatedAt,
		UpdatedAt:   kost.UpdatedAt,
		Images:      images,
	}
}
