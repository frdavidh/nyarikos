package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Room struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	KostID        uint            `json:"owner_id" gorm:"index;not null"`
	RoomType      string          `json:"room_type" gorm:"not null"`
	PricePerMonth decimal.Decimal `json:"price_per_month" gorm:"index;type:decimal(12,2);not null"`
	TotalRooms    int             `json:"total_rooms" gorm:"not null;default:1"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"-" gorm:"index" `

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

type RoomFacility struct {
	RoomID     uint `json:"room_id" gorm:"primaryKey"`
	FacilityID uint `json:"facility_id" gorm:"primaryKey"`

	Room     Room     `json:"-" gorm:"foreignKey:RoomID"`
	Facility Facility `json:"-" gorm:"foreignKey:FacilityID"`
}
