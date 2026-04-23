package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type RoomRequest struct {
	KostID        uint            `json:"kost_id" binding:"required" example:"1"`
	RoomType      string          `json:"room_type" binding:"required" example:"single"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"required" swaggertype:"string" example:"1500000.00"`
	TotalRooms    int             `json:"total_rooms" binding:"required,gt=0" example:"5"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

type RoomResponse struct {
	ID            uint               `json:"id" example:"1"`
	KostID        uint               `json:"kost_id" example:"1"`
	RoomType      string             `json:"room_type" example:"single"`
	PricePerMonth decimal.Decimal    `json:"price_per_month" swaggertype:"string" example:"1500000.00"`
	TotalRooms    int                `json:"total_rooms" example:"5"`
	CreatedAt     time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Facilities    []FacilityResponse `json:"facilities,omitempty"`
	Images        []ImageResponse    `json:"images,omitempty"`
}

type CreateRoomRequest struct {
	RoomType      string          `json:"room_type" binding:"required" example:"single"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"required" swaggertype:"string" example:"1500000.00"`
	TotalRooms    int             `json:"total_rooms" binding:"required,gt=0" example:"5"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

type UpdateRoomRequest struct {
	RoomType      string          `json:"room_type" binding:"omitempty" example:"double"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"omitempty" swaggertype:"string" example:"2000000.00"`
	TotalRooms    int             `json:"total_rooms" binding:"omitempty,gt=0" example:"3"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

// ############################################################################################################
type FacilityResponse struct {
	ID      uint    `json:"id" example:"1"`
	Name    string  `json:"name" example:"WiFi"`
	IconURL *string `json:"icon_url,omitempty" example:"https://example.com/wifi-icon.svg"`
}

type CreateFacilityRequest struct {
	Name    string  `json:"name" binding:"required" example:"AC"`
	IconURL *string `json:"icon_url" binding:"omitempty" example:"https://example.com/ac-icon.svg"`
}

type UpdateFacilityRequest struct {
	Name    string  `json:"name" binding:"omitempty" example:"Air Conditioning"`
	IconURL *string `json:"icon_url" binding:"omitempty" example:"https://example.com/ac-icon.svg"`
}

// ############################################################################################################
type CreateRoomFacilityRequest struct {
	RoomID     uint `json:"room_id" binding:"required" example:"1"`
	FacilityID uint `json:"facility_id" binding:"required" example:"1"`
}

type DeleteRoomFacilityRequest struct {
	RoomID     uint `json:"room_id" binding:"required" example:"1"`
	FacilityID uint `json:"facility_id" binding:"required" example:"1"`
}

type RoomFacilityResponse struct {
	RoomID     uint `json:"room_id" example:"1"`
	FacilityID uint `json:"facility_id" example:"1"`
}
