package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookingHandler_CreateBooking_Success(t *testing.T) {
	router := setupTestRouter()
	mockBooking := new(mockBookingService)
	handler := NewBookingHandler(mockBooking)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreateBookingRequest{
		RoomID:    1,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockBooking.On("CreateBooking", mock.Anything, uint(1), &reqBody).Return(&dto.BookingResponse{
		ID:              1,
		RoomID:          1,
		UserID:          1,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		DurationsMonths: 2,
		Status:          "pending",
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/booking/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "booking created successfully", resp["message"])
	mockBooking.AssertExpectations(t)
}

func TestBookingHandler_CreateBooking_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockBooking := new(mockBookingService)
	handler := NewBookingHandler(mockBooking)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	w := makeRequest(t, router, "POST", "/api/v1/booking/", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp["success"].(bool))
}

func TestBookingHandler_CreateBooking_RoomNotFound(t *testing.T) {
	router := setupTestRouter()
	mockBooking := new(mockBookingService)
	handler := NewBookingHandler(mockBooking)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreateBookingRequest{
		RoomID:    999,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockBooking.On("CreateBooking", mock.Anything, uint(1), &reqBody).Return(nil, services.ErrRoomNotFound)

	w := makeRequest(t, router, "POST", "/api/v1/booking/", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockBooking.AssertExpectations(t)
}

func TestBookingHandler_CreateBooking_NoRoomsAvailable(t *testing.T) {
	router := setupTestRouter()
	mockBooking := new(mockBookingService)
	handler := NewBookingHandler(mockBooking)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreateBookingRequest{
		RoomID:    1,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockBooking.On("CreateBooking", mock.Anything, uint(1), &reqBody).Return(nil, services.ErrNoRoomsAvailable)

	w := makeRequest(t, router, "POST", "/api/v1/booking/", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "no available rooms", resp["message"])
	mockBooking.AssertExpectations(t)
}

func TestBookingHandler_CreateBooking_GenericError(t *testing.T) {
	router := setupTestRouter()
	mockBooking := new(mockBookingService)
	handler := NewBookingHandler(mockBooking)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreateBookingRequest{
		RoomID:    1,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockBooking.On("CreateBooking", mock.Anything, uint(1), &reqBody).Return(nil, assert.AnError)

	w := makeRequest(t, router, "POST", "/api/v1/booking/", reqBody, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "failed to create booking", resp["message"])
	mockBooking.AssertExpectations(t)
}
