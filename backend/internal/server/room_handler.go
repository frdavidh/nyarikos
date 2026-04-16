package server

import (
	"errors"
	"net/http"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomService services.RoomService
}

func NewRoomHandler(roomService services.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	room := api.Group("/kost/:id/room")
	room.Use(middlewares...)
	room.POST("/", h.CreateRoom)
	room.GET("/", h.GetRoomByKostID)
	room.GET("/:room_id", h.GetRoomByID)
	room.PUT("/:room_id", h.UpdateRoom)
	room.DELETE("/:room_id", h.DeleteRoom)

	facility := api.Group("/facilities")
	facility.Use(middlewares...)
	facility.GET("/", h.GetAllFacilities)
	facility.POST("/", h.CreateFacility)
	facility.PUT("/:facility_id", h.UpdateFacility)
	facility.DELETE("/:facility_id", h.DeleteFacility)

	roomFacility := api.Group("/room-facilities")
	roomFacility.Use(middlewares...)
	roomFacility.POST("/", h.CreateRoomFacility)
	roomFacility.DELETE("/", h.DeleteRoomFacility)
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	kostID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	room, err := h.roomService.CreateRoom(kostID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create room", err)
		return
	}

	utils.CreatedResponse(c, "Room created successfully", room)
}

func (h *RoomHandler) GetRoomByKostID(c *gin.Context) {
	kostID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	rooms, err := h.roomService.GetRoomByKostID(kostID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get rooms", err)
		return
	}

	utils.SuccessResponse(c, "Rooms retrieved successfully", rooms)
}

func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		return
	}

	room, err := h.roomService.GetRoomByID(roomID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "Failed to get room", err)
		}
		return
	}

	utils.SuccessResponse(c, "Room retrieved successfully", room)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		return
	}

	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	room, err := h.roomService.UpdateRoom(roomID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to update room", err)
		}
		return
	}

	utils.SuccessResponse(c, "Room updated successfully", room)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		return
	}

	if err := h.roomService.DeleteRoom(roomID); err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to delete room", err)
		}
		return
	}

	utils.SuccessResponse(c, "Room deleted successfully", nil)
}

// ############################################################################################################
func (h *RoomHandler) GetAllFacilities(c *gin.Context) {
	facilities, err := h.roomService.GetAllFacilities()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get facilities", err)
		return
	}
	utils.SuccessResponse(c, "Facilities retrieved successfully", facilities)
}

func (h *RoomHandler) CreateFacility(c *gin.Context) {
	var req dto.CreateFacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	facility, err := h.roomService.CreateFacility(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create facility", err)
		return
	}

	utils.CreatedResponse(c, "Facility created successfully", facility)
}

func (h *RoomHandler) UpdateFacility(c *gin.Context) {
	facilityID, ok := parseUintParam(c, "facility_id")
	if !ok {
		return
	}

	var req dto.UpdateFacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	facility, err := h.roomService.UpdateFacility(facilityID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrFacilityNotFound):
			utils.NotFoundResponse(c, "facility not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to update facility", err)
		}
		return
	}

	utils.SuccessResponse(c, "Facility updated successfully", facility)
}

func (h *RoomHandler) DeleteFacility(c *gin.Context) {
	facilityID, ok := parseUintParam(c, "facility_id")
	if !ok {
		return
	}

	if err := h.roomService.DeleteFacility(facilityID); err != nil {
		switch {
		case errors.Is(err, services.ErrFacilityNotFound):
			utils.NotFoundResponse(c, "facility not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to delete facility", err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility deleted successfully"})
}

// ############################################################################################################
func (h *RoomHandler) CreateRoomFacility(c *gin.Context) {
	var req dto.CreateRoomFacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	roomFacility, err := h.roomService.CreateRoomFacility(req.RoomID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to create room facility", err)
		}
		return
	}

	utils.CreatedResponse(c, "Room facility created successfully", roomFacility)
}

func (h *RoomHandler) DeleteRoomFacility(c *gin.Context) {
	var req dto.DeleteRoomFacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	if err := h.roomService.DeleteRoomFacility(req.RoomID, req.FacilityID); err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		case errors.Is(err, services.ErrFacilityNotFound):
			utils.NotFoundResponse(c, "facility not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to delete room facility", err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room facility deleted successfully"})
}
