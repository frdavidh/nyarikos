package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null"`
	Password    *string        `json:"-"`
	PhoneNumber *string        `json:"phone_number" gorm:"uniqueIndex"`
	Role        UserRole       `json:"role" gorm:"type:user_role;default:pencari"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Kosts          []Kost          `json:"kosts,omitempty" gorm:"foreignKey:OwnerID"`
	RefreshTokens  []RefreshToken  `json:"-"`
	SocialAccounts []SocialAccount `json:"-"`
	Bookings       []Booking       `json:"bookings,omitempty"`
}

type RefreshToken struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	Token      string    `json:"token" gorm:"uniqueIndex;not null"`
	DeviceInfo *string   `json:"device_info"`
	IPAddress  *string   `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	IsRevoked  *bool     `json:"is_revoked" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`

	User User `json:"-" gorm:"foreignKey:UserID"`
}

type SocialAccount struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	UserID       uint          `json:"user_id" gorm:"not null"`
	ProviderName OAuthProvider `json:"provider" gorm:"type:oauth_provider;not null"`
	ProviderID   string        `json:"provider_id" gorm:"uniqueIndex;not null"`
	CreatedAt    time.Time     `json:"created_at"`

	User User `json:"-" gorm:"foreignKey:UserID" `
}

type UserRole string

const (
	RolePencari UserRole = "pencari"
	RolePemilik UserRole = "pemilik"
	RoleAdmin   UserRole = "admin"
)

type OAuthProvider string

const (
	ProviderGoogle   OAuthProvider = "google"
	ProviderFacebook OAuthProvider = "facebook"
)
