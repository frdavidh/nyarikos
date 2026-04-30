package services

import (
	"errors"
	"fmt"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"gorm.io/gorm"
)

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrFacilityNotFound = errors.New("facility not found")
)

type RoomService interface {
	CreateRoom(kostID uint, req *dto.CreateRoomRequest) (*dto.RoomResponse, error)
	UpdateRoom(roomID uint, req *dto.UpdateRoomRequest) (*dto.RoomResponse, error)
	DeleteRoom(roomID uint) error
	GetRoomByID(roomID uint) (*dto.RoomResponse, error)
	GetRoomByKostID(kostID uint) ([]dto.RoomResponse, error)

	GetAllFacilities() ([]dto.FacilityResponse, error)
	CreateFacility(req *dto.CreateFacilityRequest) (*dto.FacilityResponse, error)
	UpdateFacility(facilityID uint, req *dto.UpdateFacilityRequest) (*dto.FacilityResponse, error)
	DeleteFacility(facilityID uint) error

	CreateRoomFacility(roomID uint, req *dto.CreateRoomFacilityRequest) (*dto.RoomFacilityResponse, error)
	DeleteRoomFacility(roomID uint, facilityID uint) error
}

type roomService struct {
	db *gorm.DB
}

func NewRoomService(db *gorm.DB) RoomService {
	return &roomService{db: db}
}

func (s *roomService) CreateRoom(kostID uint, req *dto.CreateRoomRequest) (*dto.RoomResponse, error) {
	room := models.Room{
		KostID:        kostID,
		RoomType:      req.RoomType,
		PricePerMonth: req.PricePerMonth,
		TotalRooms:    req.TotalRooms,
	}

	if err := s.db.Create(&room).Error; err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	if len(req.FacilityIDs) > 0 {
		var facilities []models.Facility
		if err := s.db.Where("id IN ?", req.FacilityIDs).Find(&facilities).Error; err != nil {
			return nil, fmt.Errorf("failed to find facilities: %w", err)
		}
		if err := s.db.Model(&room).Association("Facilities").Replace(facilities); err != nil {
			return nil, fmt.Errorf("failed to associate facilities: %w", err)
		}
		room.Facilities = facilities
	}

	return toRoomResponse(&room), nil
}

func (s *roomService) UpdateRoom(roomID uint, req *dto.UpdateRoomRequest) (*dto.RoomResponse, error) {
	var room models.Room
	if err := s.db.Preload("Facilities").First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	if req.RoomType != "" {
		room.RoomType = req.RoomType
	}
	if !req.PricePerMonth.IsZero() {
		room.PricePerMonth = req.PricePerMonth
	}
	if req.TotalRooms > 0 {
		room.TotalRooms = req.TotalRooms
	}

	if err := s.db.Save(&room).Error; err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	if req.FacilityIDs != nil {
		var facilities []models.Facility
		if len(req.FacilityIDs) > 0 {
			if err := s.db.Where("id IN ?", req.FacilityIDs).Find(&facilities).Error; err != nil {
				return nil, fmt.Errorf("failed to find facilities: %w", err)
			}
		}
		if err := s.db.Model(&room).Association("Facilities").Replace(facilities); err != nil {
			return nil, fmt.Errorf("failed to sync facilities: %w", err)
		}
		room.Facilities = facilities
	}

	return toRoomResponse(&room), nil
}

func (s *roomService) DeleteRoom(roomID uint) error {
	result := s.db.Delete(&models.Room{}, roomID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete room: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRoomNotFound
	}
	return nil
}

func (s *roomService) GetRoomByID(roomID uint) (*dto.RoomResponse, error) {
	var room models.Room
	if err := s.db.Preload("Facilities").First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	return toRoomResponse(&room), nil
}

func (s *roomService) GetRoomByKostID(kostID uint) ([]dto.RoomResponse, error) {
	var rooms []models.Room
	if err := s.db.Preload("Facilities").Where("kost_id = ?", kostID).Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	responses := make([]dto.RoomResponse, len(rooms))
	for i := range rooms {
		responses[i] = *toRoomResponse(&rooms[i])
	}
	return responses, nil
}

func (s *roomService) GetAllFacilities() ([]dto.FacilityResponse, error) {
	var facilities []models.Facility
	if err := s.db.Find(&facilities).Error; err != nil {
		return nil, fmt.Errorf("failed to get facilities: %w", err)
	}
	responses := make([]dto.FacilityResponse, len(facilities))
	for i, f := range facilities {
		responses[i] = toFacilityResponse(&f)
	}
	return responses, nil
}

func (s *roomService) CreateFacility(req *dto.CreateFacilityRequest) (*dto.FacilityResponse, error) {
	facility := models.Facility{
		Name:    req.Name,
		IconURL: req.IconURL,
	}
	if err := s.db.Create(&facility).Error; err != nil {
		return nil, fmt.Errorf("failed to create facility: %w", err)
	}
	res := toFacilityResponse(&facility)
	return &res, nil
}

func (s *roomService) UpdateFacility(facilityID uint, req *dto.UpdateFacilityRequest) (*dto.FacilityResponse, error) {
	var facility models.Facility
	if err := s.db.First(&facility, facilityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFacilityNotFound
		}
		return nil, fmt.Errorf("failed to get facility: %w", err)
	}

	if req.Name != "" {
		facility.Name = req.Name
	}
	if req.IconURL != nil {
		facility.IconURL = req.IconURL
	}

	if err := s.db.Save(&facility).Error; err != nil {
		return nil, fmt.Errorf("failed to update facility: %w", err)
	}
	res := toFacilityResponse(&facility)
	return &res, nil
}

func (s *roomService) DeleteFacility(facilityID uint) error {
	result := s.db.Delete(&models.Facility{}, facilityID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete facility: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFacilityNotFound
	}
	return nil
}

func (s *roomService) CreateRoomFacility(roomID uint, req *dto.CreateRoomFacilityRequest) (*dto.RoomFacilityResponse, error) {
	var room models.Room
	if err := s.db.First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	var facility models.Facility
	if err := s.db.First(&facility, req.FacilityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFacilityNotFound
		}
		return nil, fmt.Errorf("failed to get facility: %w", err)
	}

	roomFacility := models.RoomFacility{
		RoomID:     room.ID,
		FacilityID: facility.ID,
	}

	if err := s.db.Create(&roomFacility).Error; err != nil {
		return nil, fmt.Errorf("failed to create room facility: %w", err)
	}

	return toRoomFacilityResponse(&roomFacility), nil
}

func (s *roomService) DeleteRoomFacility(roomID uint, facilityID uint) error {
	result := s.db.Where("room_id = ? AND facility_id = ?", roomID, facilityID).Delete(&models.RoomFacility{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete room facility: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRoomNotFound
	}
	return nil
}

func toRoomResponse(room *models.Room) *dto.RoomResponse {
	resp := &dto.RoomResponse{
		ID:            room.ID,
		KostID:        room.KostID,
		RoomType:      room.RoomType,
		PricePerMonth: room.PricePerMonth,
		TotalRooms:    room.TotalRooms,
		CreatedAt:     room.CreatedAt,
		UpdatedAt:     room.UpdatedAt,
	}
	for _, f := range room.Facilities {
		resp.Facilities = append(resp.Facilities, toFacilityResponse(&f))
	}
	return resp
}

func toFacilityResponse(f *models.Facility) dto.FacilityResponse {
	return dto.FacilityResponse{
		ID:      f.ID,
		Name:    f.Name,
		IconURL: f.IconURL,
	}
}

func toRoomFacilityResponse(rf *models.RoomFacility) *dto.RoomFacilityResponse {
	return &dto.RoomFacilityResponse{
		RoomID:     rf.RoomID,
		FacilityID: rf.FacilityID,
	}
}
