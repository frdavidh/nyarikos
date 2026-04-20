package services

import (
	"errors"
	"fmt"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"gorm.io/gorm"
)

var (
	ErrBookingNotFound  = errors.New("booking not found")
	ErrNoRoomsAvailable = errors.New("no available rooms")
)

type BookingService interface {
	CreateBooking(userID uint, req *dto.CreateBookingRequest) (*dto.BookingResponse, error)
}

type bookingService struct {
	db *gorm.DB
}

func NewBookingService(db *gorm.DB) BookingService {
	return &bookingService{db: db}
}

func (s *bookingService) CreateBooking(userID uint, req *dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	if !req.EndDate.After(req.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	years := req.EndDate.Year() - req.StartDate.Year()
	months := int(req.EndDate.Month()) - int(req.StartDate.Month())
	durationsMonths := years*12 + months
	if durationsMonths < 1 {
		return nil, fmt.Errorf("booking must be at least 1 month")
	}

	var booking models.Booking

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var room models.Room
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&room, req.RoomID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoomNotFound
			}
			return fmt.Errorf("failed to get room: %w", err)
		}

		var activeBookings int64
		tx.Model(&models.Booking{}).
			Where("room_id = ? AND status IN ?", req.RoomID, []models.BookingStatus{models.BookingPending, models.BookingPaid}).
			Count(&activeBookings)

		if activeBookings >= int64(room.TotalRooms) {
			return ErrNoRoomsAvailable
		}

		booking = models.Booking{
			UserID:          userID,
			RoomID:          req.RoomID,
			StartDate:       req.StartDate,
			EndDate:         req.EndDate,
			DurationsMonths: durationsMonths,
			Status:          models.BookingPending,
		}

		if err := tx.Create(&booking).Error; err != nil {
			return fmt.Errorf("failed to create booking: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return toBookingResponse(&booking), nil
}

func toBookingResponse(b *models.Booking) *dto.BookingResponse {
	var totalPrice float64
	if b.TotalPrice != nil {
		totalPrice = *b.TotalPrice
	}
	return &dto.BookingResponse{
		ID:              b.ID,
		BookingCode:     b.BookingCode,
		RoomID:          b.RoomID,
		UserID:          b.UserID,
		StartDate:       b.StartDate,
		EndDate:         b.EndDate,
		DurationsMonths: b.DurationsMonths,
		TotalPrice:      totalPrice,
		Status:          string(b.Status),
		CreatedAt:       b.CreatedAt,
	}
}
