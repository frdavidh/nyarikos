package models

import "time"

type Booking struct {
	ID              uint          `json:"id" gorm:"primaryKey" `
	BookingCode     string        `json:"booking_code" gorm:"uniqueIndex;not null" `
	UserID          uint          `json:"user_id" gorm:"index;not null" `
	RoomID          uint          `json:"room_id" gorm:"not null" `
	StartDate       time.Time     `json:"start_date" gorm:"type:date;not null" `
	EndDate         time.Time     `json:"end_date" gorm:"type:date;not null" `
	DurationsMonths int           `json:"durations_months" gorm:"not null;default:1" `
	TotalPrice      *float64      `json:"total_price" gorm:"type:decimal(12,2)" `
	Status          BookingStatus `json:"status" gorm:"type:booking_status;default:'pending'" `
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`

	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID" `
	Room    Room     `json:"room,omitempty" gorm:"foreignKey:RoomID" `
	Payment *Payment `json:"payment,omitempty" gorm:"foreignKey:BookingID" `
}

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingPaid      BookingStatus = "paid"
	BookingCancelled BookingStatus = "cancelled"
	BookingCompleted BookingStatus = "completed"
)
