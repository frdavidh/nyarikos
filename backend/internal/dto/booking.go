package dto

import "time"

type CreateBookingRequest struct {
	RoomID    uint      `json:"room_id" binding:"required" example:"1"`
	StartDate time.Time `json:"start_date" binding:"required" time_format:"2006-01-02" example:"2024-06-01"`
	EndDate   time.Time `json:"end_date" binding:"required" time_format:"2006-01-02" example:"2024-09-01"`
}

type BookingResponse struct {
	ID              uint      `json:"id" example:"1"`
	BookingCode     string    `json:"booking_code" example:"BK-20240601-ABC123"`
	RoomID          uint      `json:"room_id" example:"1"`
	UserID          uint      `json:"user_id" example:"1"`
	StartDate       time.Time `json:"start_date" example:"2024-06-01T00:00:00Z"`
	EndDate         time.Time `json:"end_date" example:"2024-09-01T00:00:00Z"`
	DurationsMonths int       `json:"durations_months" example:"3"`
	TotalPrice      float64   `json:"total_price" example:"4500000.00"`
	Status          string    `json:"status" example:"pending"`
	CreatedAt       time.Time `json:"created_at" example:"2024-05-15T10:30:00Z"`
}
