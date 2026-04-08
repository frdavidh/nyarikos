package services

import (
	"errors"
	"fmt"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrKostNotFound = errors.New("record not found")
)

type KostService interface {
	CreateKost(ownerID uint, req *dto.CreateKostRequest) (*dto.KostResponse, error)
	UpdateKost(kostID, userID uint, req *dto.UpdateKostRequest) (*dto.KostResponse, error)
	DeleteKost(id uint) (*dto.KostResponse, error)
	// GetKost(req *dto.GetKostRequest) (*dto.KostResponse, error)
	GetAllKost(page, limit int) ([]dto.KostResponse, int64, error)
	GetKost(id uint) (*dto.KostResponse, error)
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
	}

	if err := s.db.Create(&kost).Error; err != nil {
		return nil, fmt.Errorf("failed to create kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

// func isUniqueContaintError(err error) bool {
// 	var pgErr *pgconn.PgError
// 	if errors.As(err, &pgErr) {
// 		return pgErr.Code == "23505"
// 	}
// 	return false
// }

func (s *kostService) GetAllKost(page, limit int) ([]dto.KostResponse, int64, error) {
	var kosts []models.Kost
	var total int64

	if err := s.db.Model(&models.Kost{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count kost: %w", err)
	}

	offset := (page - 1) * limit
	if err := s.db.Offset(offset).Limit(limit).Find(&kosts).Error; err != nil {
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
	if err := s.db.First(&kost, id).Error; err != nil {
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

	if err := s.db.Save(&kost).Error; err != nil {
		return nil, fmt.Errorf("failed to update kost: %w", err)
	}
	return toKostResponse(&kost), nil
}

func (s *kostService) DeleteKost(userID uint) (*dto.KostResponse, error) {
	var kost models.Kost
	if err := s.db.First(&kost, userID).Error; err != nil {
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
	// return &dto.KostResponse{
	// 	ID:          kost.ID,
	// 	OwnerID:     kost.OwnerID,
	// 	Name:        kost.Name,
	// 	Description: kost.Description,
	// 	Address:     kost.Address,
	// 	City:        kost.City,
	// 	IsPremium:   kost.IsPremium,
	// 	CreatedAt:   kost.CreatedAt,
	// 	UpdatedAt:   kost.UpdatedAt,
	// }, nil
	return toKostResponse(&kost), nil
}

func toKostResponse(kost *models.Kost) *dto.KostResponse {
	return &dto.KostResponse{
		ID:          kost.ID,
		OwnerID:     kost.OwnerID,
		Name:        kost.Name,
		Description: kost.Description,
		Address:     kost.Address,
		City:        kost.City,
		IsPremium:   kost.IsPremium,
		CreatedAt:   kost.CreatedAt,
		UpdatedAt:   kost.UpdatedAt,
	}
}
