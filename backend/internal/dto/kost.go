package dto

import "time"

type ImageResponse struct {
	ID       uint   `json:"id"`
	ImageURL string `json:"image_url"`
	IsMain   bool   `json:"is_main"`
}
type KostRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description" binding:"omitempty"`
	Address     string `json:"address" binding:"required"`
	City        string `json:"city" binding:"required"`
	IsPremium   *bool  `json:"is_premium" binding:"omitempty"`
}

type KostResponse struct {
	ID          uint            `json:"id"`
	OwnerID     uint            `json:"owner_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Address     string          `json:"address"`
	City        string          `json:"city"`
	IsPremium   bool            `json:"is_premium"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Images      []ImageResponse `json:"images,omitempty"`
}

type CreateKostRequest struct {
	OwnerID     uint   `json:"owner_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description" binding:"omitempty"`
	Address     string `json:"address" binding:"required"`
	City        string `json:"city" binding:"required"`
	IsPremium   *bool  `json:"is_premium" binding:"omitempty"`
}

type UpdateKostRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty"`
	Address     *string `json:"address" binding:"omitempty"`
	City        *string `json:"city" binding:"omitempty"`
	IsPremium   *bool   `json:"is_premium" binding:"omitempty"`
}
