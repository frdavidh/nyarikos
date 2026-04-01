package dto

import "time"

type RoomRequest struct {
	KostID        uint    `json:"kost_id" binding:"required"`
	RoomType      string  `json:"room_type" binding:"required"`
	PricePerMonth float64 `json:"price_per_month" binding:"required,gt=0"`
	TotalRooms    int     `json:"total_rooms" binding:"required,gt=0"`
	FacilityIDs   []uint  `json:"facility_ids" binding:"omitempty"`
}

type RoomResponse struct {
	ID            uint               `json:"id"`
	KostID        uint               `json:"kost_id"`
	RoomType      string             `json:"room_type"`
	PricePerMonth float64            `json:"price_per_month"`
	TotalRooms    int                `json:"total_rooms"`
	CreatedAt     time.Time          `json:"created_at"`
	Facilities    []FacilityResponse `json:"facilities,omitempty"`
	Images        []ImageResponse    `json:"images,omitempty"`
}

type FacilityResponse struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	IconURL *string `json:"icon_url,omitempty"`
}
