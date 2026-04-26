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
	GetAllKost(req *dto.SearchKostRequest) ([]dto.KostResponse, int64, error)
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

func (s *kostService) GetAllKost(req *dto.SearchKostRequest) ([]dto.KostResponse, int64, error) {
	buildQuery := func(db *gorm.DB) *gorm.DB {
		q := db.Model(&models.Kost{})

		if req.Q != "" {
			like := "%" + req.Q + "%"
			q = q.Where("LOWER(name) LIKE LOWER(?) OR LOWER(address) LIKE LOWER(?) OR LOWER(city) LIKE LOWER(?)", like, like, like)
		}
		if req.City != "" {
			q = q.Where("LOWER(city) LIKE LOWER(?)", "%"+req.City+"%")
		}
		if req.KostType != "" {
			q = q.Where("kost_type = ?", req.KostType)
		}
		if req.MinPrice > 0 || req.MaxPrice > 0 || req.RoomType != "" || len(req.FacilityIDs) > 0 {
			roomQuery := s.db.Model(&models.Room{}).Select("DISTINCT kost_id")
			if req.MinPrice > 0 {
				roomQuery = roomQuery.Where("price_per_month >= ?", req.MinPrice)
			}
			if req.MaxPrice > 0 {
				roomQuery = roomQuery.Where("price_per_month <= ?", req.MaxPrice)
			}
			if req.RoomType != "" {
				roomQuery = roomQuery.Where("room_type = ?", req.RoomType)
			}
			if len(req.FacilityIDs) > 0 {
				facilitySubQuery := s.db.Model(&models.RoomFacility{}).
					Select("room_id").
					Where("facility_id IN ?", req.FacilityIDs).
					Group("room_id").
					Having("COUNT(DISTINCT facility_id) = ?", len(req.FacilityIDs))
				roomQuery = roomQuery.Where("id IN (?)", facilitySubQuery)
			}
			q = q.Where("id IN (?)", roomQuery)
		}
		return q
	}

	var total int64
	if err := buildQuery(s.db).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count kost: %w", err)
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	var kosts []models.Kost
	if err := buildQuery(s.db).
		Preload("Images").
		Preload("Rooms").
		Preload("Rooms.Facilities").
		Offset(offset).
		Limit(req.Limit).
		Find(&kosts).Error; err != nil {
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
	if err := s.db.Preload("Images").Preload("Rooms").Preload("Rooms.Facilities").First(&kost, id).Error; err != nil {
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
	rooms := make([]dto.RoomResponse, 0, len(kost.Rooms))
	for i := range kost.Rooms {
		rooms = append(rooms, *toRoomResponse(&kost.Rooms[i]))
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
		Rooms:       rooms,
	}
}
