package models

import (
	"time"

	"gorm.io/gorm"
)

type Kost struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	OwnerID     uint           `json:"owner_id" gorm:"index;not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	Address     string         `json:"address" gorm:"not null"`
	City        string         `json:"city" gorm:"index;not null"`
	IsPremium   bool           `json:"is_premium" gorm:"default:false"`
	KostType    string         `json:"kost_type" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Owner  User        `json:"users" gorm:"foreignKey:OwnerID"`
	Rooms  []Room      `json:"rooms" gorm:"foreignKey:KostID"`
	Images []KostImage `json:"images" gorm:"foreignKey:KostID"`
}

type KostImage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	KostID    uint           `json:"kost_id" gorm:"index;not null"`
	ImageURL  string         `json:"image_url" gorm:"not null"`
	AltText   string         `json:"alt_text" gorm:"not null"`
	IsMain    bool           `json:"is_main" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
