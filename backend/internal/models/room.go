package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	KostID        uint           `json:"owner_id" gorm:"index;not null"`
	RoomType      string         `json:"name" gorm:"not null"`
	PricePerMonth float64        `json:"description" gorm:"index;type:decimal(12,2);not null"`
	TotalRooms    string         `json:"address" gorm:"not null;default:1"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index" `

	Kost       Kost       `json:"-" gorm:"foreignKey:KostID" `
	Facilities []Facility `json:"facilities,omitempty" gorm:"many2many:room_facilities;"`
	Bookings   []Booking  `json:"bookings,omitempty" gorm:"foreignKey:RoomID"`
}

type Facility struct {
	ID      uint    `json:"id" gorm:"primaryKey" `
	Name    string  `json:"name" gorm:"uniqueIndex;not null" `
	IconURL *string `json:"icon_url"`

	Rooms []Room `json:"-" gorm:"many2many:room_facilities;" `
}
