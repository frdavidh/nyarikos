package dto

import "time"

type CreateBookingRequest struct {
	RoomID          uint      `json:"room_id" binding:"required"`
	StartDate       time.Time `json:"start_date" binding:"required" time_format:"2006-01-02"`
	EndDate         time.Time `json:"end_date" binding:"required" time_format:"2006-01-02"`
	DurationsMonths int       `json:"durations_months" binding:"required,min=1"`
}

type BookingResponse struct {
	ID              uint      `json:"id"`
	BookingCode     string    `json:"booking_code"`
	RoomID          uint      `json:"room_id"`
	UserID          uint      `json:"user_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	DurationsMonths int       `json:"durations_months"`
	TotalPrice      float64   `json:"total_price"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
