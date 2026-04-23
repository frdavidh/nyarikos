package dto

import (
	"time"
)

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=6" example:"password123"`
	PhoneNumber string `json:"phone_number" binding:"omitempty" example:"+6281234567890"`
	Name        string `json:"name" example:"John Doe"`
	Role        string `json:"role" binding:"omitempty,oneof=pencari pemilik" example:"pencari"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type UserResponse struct {
	ID          uint      `json:"id" example:"1"`
	Name        string    `json:"name" example:"John Doe"`
	Email       string    `json:"email" example:"user@example.com"`
	PhoneNumber *string   `json:"phone_number,omitempty" example:"+6281234567890"`
	Role        string    `json:"role" example:"pencari"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type UpdateProfileRequest struct {
	Name        *string `json:"name" example:"John Doe"`
	PhoneNumber *string `json:"phone_number" example:"+6281234567890"`
}
