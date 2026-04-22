package services

import (
	"errors"
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBooking_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	room := models.Room{
		KostID:     1,
		RoomType:   "Standard",
		TotalRooms: 2,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	req := &dto.CreateBookingRequest{
		RoomID:    room.ID,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, room.ID, resp.RoomID)
	assert.Equal(t, uint(1), resp.UserID)
	assert.Equal(t, 2, resp.DurationsMonths)
	assert.Equal(t, "pending", resp.Status)
	assert.NotZero(t, resp.ID)
}

func TestCreateBooking_EndDateBeforeStartDate(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	req := &dto.CreateBookingRequest{
		RoomID:    1,
		StartDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "end_date must be after start_date")
}

func TestCreateBooking_DurationLessThanOneMonth(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	req := &dto.CreateBookingRequest{
		RoomID:    1,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "booking must be at least 1 month")
}

func TestCreateBooking_RoomNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	req := &dto.CreateBookingRequest{
		RoomID:    999,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrRoomNotFound))
}

func TestCreateBooking_NoRoomsAvailable(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	room := models.Room{
		KostID:     1,
		RoomType:   "Standard",
		TotalRooms: 1,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	existingBooking := models.Booking{
		BookingCode:     "BK001",
		UserID:          2,
		RoomID:          room.ID,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		DurationsMonths: 5,
		Status:          models.BookingPaid,
	}
	err = db.Create(&existingBooking).Error
	require.NoError(t, err)

	req := &dto.CreateBookingRequest{
		RoomID:    room.ID,
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrNoRoomsAvailable))
}

func TestCreateBooking_CancelledBookingDoesNotBlock(t *testing.T) {
	db := setupTestDB(t)
	service := NewBookingService(db)

	room := models.Room{
		KostID:     1,
		RoomType:   "Standard",
		TotalRooms: 1,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	cancelledBooking := models.Booking{
		BookingCode:     "BK001",
		UserID:          2,
		RoomID:          room.ID,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		DurationsMonths: 5,
		Status:          models.BookingCancelled,
	}
	err = db.Create(&cancelledBooking).Error
	require.NoError(t, err)

	req := &dto.CreateBookingRequest{
		RoomID:    room.ID,
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	resp, err := service.CreateBooking(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestToBookingResponse_WithTotalPrice(t *testing.T) {
	price := 1500.50
	booking := &models.Booking{
		ID:              1,
		BookingCode:     "BK001",
		RoomID:          2,
		UserID:          3,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		DurationsMonths: 2,
		TotalPrice:      &price,
		Status:          models.BookingPaid,
	}

	resp := toBookingResponse(booking)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, "BK001", resp.BookingCode)
	assert.Equal(t, uint(2), resp.RoomID)
	assert.Equal(t, uint(3), resp.UserID)
	assert.Equal(t, 2, resp.DurationsMonths)
	assert.Equal(t, 1500.50, resp.TotalPrice)
	assert.Equal(t, "paid", resp.Status)
}

func TestToBookingResponse_WithoutTotalPrice(t *testing.T) {
	booking := &models.Booking{
		ID:              1,
		BookingCode:     "BK002",
		RoomID:          2,
		UserID:          3,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		DurationsMonths: 2,
		Status:          models.BookingPending,
	}

	resp := toBookingResponse(booking)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, 0.0, resp.TotalPrice)
}
