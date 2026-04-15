package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type RoomRequest struct {
	KostID        uint            `json:"kost_id" binding:"required"`
	RoomType      string          `json:"room_type" binding:"required"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"required"`
	TotalRooms    int             `json:"total_rooms" binding:"required,gt=0"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

type RoomResponse struct {
	ID            uint               `json:"id"`
	KostID        uint               `json:"kost_id"`
	RoomType      string             `json:"room_type"`
	PricePerMonth decimal.Decimal    `json:"price_per_month"`
	TotalRooms    int                `json:"total_rooms"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Facilities    []FacilityResponse `json:"facilities,omitempty"`
	Images        []ImageResponse    `json:"images,omitempty"`
}

type CreateRoomRequest struct {
	RoomType      string          `json:"room_type" binding:"required"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"required"`
	TotalRooms    int             `json:"total_rooms" binding:"required,gt=0"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

type UpdateRoomRequest struct {
	RoomType      string          `json:"room_type" binding:"omitempty"`
	PricePerMonth decimal.Decimal `json:"price_per_month" binding:"omitempty"`
	TotalRooms    int             `json:"total_rooms" binding:"omitempty,gt=0"`
	FacilityIDs   []uint          `json:"facility_ids" binding:"omitempty"`
}

// ############################################################################################################
type FacilityResponse struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	IconURL *string `json:"icon_url,omitempty"`
}

type CreateFacilityRequest struct {
	Name    string  `json:"name" binding:"required"`
	IconURL *string `json:"icon_url" binding:"omitempty"`
}

type UpdateFacilityRequest struct {
	Name    string  `json:"name" binding:"omitempty"`
	IconURL *string `json:"icon_url" binding:"omitempty"`
}

// ############################################################################################################
type CreateRoomFacilityRequest struct {
	RoomID     uint `json:"room_id" binding:"required"`
	FacilityID uint `json:"facility_id" binding:"required"`
}

type DeleteRoomFacilityRequest struct {
	RoomID     uint `json:"room_id" binding:"required"`
	FacilityID uint `json:"facility_id" binding:"required"`
}

type RoomFacilityResponse struct {
	RoomID     uint `json:"room_id"`
	FacilityID uint `json:"facility_id"`
}
