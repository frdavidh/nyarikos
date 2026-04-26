package dto

import "time"

type ImageResponse struct {
	ID       uint   `json:"id" example:"1"`
	ImageURL string `json:"image_url" example:"https://example.com/image.jpg"`
	IsMain   bool   `json:"is_main" example:"true"`
}

type KostResponse struct {
	ID          uint            `json:"id" example:"1"`
	OwnerID     uint            `json:"owner_id" example:"1"`
	Name        string          `json:"name" example:"Kost Bahagia"`
	Description *string         `json:"description,omitempty" example:"A comfortable kost"`
	Address     string          `json:"address" example:"Jl. Merdeka No. 1"`
	City        string          `json:"city" example:"Jakarta"`
	IsPremium   bool            `json:"is_premium" example:"true"`
	CreatedAt   time.Time       `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time       `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Images      []ImageResponse `json:"images,omitempty"`
	KostType    string          `json:"kost_type" example:"putra"`
	Rooms       []RoomResponse  `json:"rooms,omitempty"`
}

type CreateKostRequest struct {
	Name        string `json:"name" binding:"required,max=255" example:"Kost Bahagia"`
	Description string `json:"description" binding:"omitempty" example:"A comfortable kost near the university"`
	Address     string `json:"address" binding:"required" example:"Jl. Merdeka No. 1"`
	City        string `json:"city" binding:"required" example:"Jakarta"`
	IsPremium   *bool  `json:"is_premium" binding:"omitempty" example:"true"`
	KostType    string `json:"kost_type" binding:"required" example:"putra"`
}

type SearchKostRequest struct {
	Q           string
	MinPrice    float64
	MaxPrice    float64
	RoomType    string
	FacilityIDs []uint
	City        string
	KostType    string
	Page        int
	Limit       int
}

type UpdateKostRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=255" example:"Kost Bahagia"`
	Description *string `json:"description" binding:"omitempty" example:"Updated description"`
	Address     *string `json:"address" binding:"omitempty" example:"Jl. Merdeka No. 2"`
	City        *string `json:"city" binding:"omitempty" example:"Jakarta"`
	IsPremium   *bool   `json:"is_premium" binding:"omitempty" example:"false"`
	KostType    *string `json:"kost_type" binding:"omitempty" example:"putri"`
}
