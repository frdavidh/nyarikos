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

// @Tags			Room
// @Summary		Create a new room
// @Description	Create a new room for a specific kost
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int										true	"Kost ID"
// @Param			request	body		dto.CreateRoomRequest					true	"Create Room Request"
// @Success		201		{object}	utils.Response{data=dto.RoomResponse}	"Room created successfully"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/kost/{id}/room/ [post]
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

// @Tags			Room
// @Summary		Get rooms by kost ID
// @Description	Get all rooms belonging to a specific kost
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		int										true	"Kost ID"
// @Success		200	{object}	utils.Response{data=[]dto.RoomResponse}	"Rooms retrieved successfully"
// @Failure		401	{object}	utils.Response							"Unauthorized"
// @Failure		500	{object}	utils.Response							"Internal server error"
// @Router			/kost/{id}/room/ [get]
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

// @Tags			Room
// @Summary		Get room by ID
// @Description	Get detailed information about a specific room
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int										true	"Kost ID"
// @Param			room_id	path		int										true	"Room ID"
// @Success		200		{object}	utils.Response{data=dto.RoomResponse}	"Room retrieved successfully"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		404		{object}	utils.Response							"Room not found"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/kost/{id}/room/{room_id} [get]
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

// @Tags			Room
// @Summary		Update room
// @Description	Update a room's information
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int										true	"Kost ID"
// @Param			room_id	path		int										true	"Room ID"
// @Param			request	body		dto.UpdateRoomRequest					true	"Update Room Request"
// @Success		200		{object}	utils.Response{data=dto.RoomResponse}	"Room updated successfully"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		404		{object}	utils.Response							"Room not found"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/kost/{id}/room/{room_id} [put]
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

// @Tags			Room
// @Summary		Delete room
// @Description	Delete a room
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int				true	"Kost ID"
// @Param			room_id	path		int				true	"Room ID"
// @Success		200		{object}	utils.Response	"Room deleted successfully"
// @Failure		401		{object}	utils.Response	"Unauthorized"
// @Failure		404		{object}	utils.Response	"Room not found"
// @Failure		500		{object}	utils.Response	"Internal server error"
// @Router			/kost/{id}/room/{room_id} [delete]
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
//
//	@Tags			Facility
//	@Summary		Get all facilities
//	@Description	Get a list of all available facilities
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.Response{data=[]dto.FacilityResponse}	"Facilities retrieved successfully"
//	@Failure		401	{object}	utils.Response								"Unauthorized"
//	@Failure		500	{object}	utils.Response								"Internal server error"
//	@Router			/facilities/ [get]
func (h *RoomHandler) GetAllFacilities(c *gin.Context) {
	facilities, err := h.roomService.GetAllFacilities()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get facilities", err)
		return
	}
	utils.SuccessResponse(c, "Facilities retrieved successfully", facilities)
}

// @Tags			Facility
// @Summary		Create a new facility
// @Description	Create a new facility type
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.CreateFacilityRequest					true	"Create Facility Request"
// @Success		201		{object}	utils.Response{data=dto.FacilityResponse}	"Facility created successfully"
// @Failure		400		{object}	utils.Response								"Invalid request"
// @Failure		401		{object}	utils.Response								"Unauthorized"
// @Failure		500		{object}	utils.Response								"Internal server error"
// @Router			/facilities/ [post]
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

// @Tags			Facility
// @Summary		Update facility
// @Description	Update a facility type
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			facility_id	path		int											true	"Facility ID"
// @Param			request		body		dto.UpdateFacilityRequest					true	"Update Facility Request"
// @Success		200			{object}	utils.Response{data=dto.FacilityResponse}	"Facility updated successfully"
// @Failure		400			{object}	utils.Response								"Invalid request"
// @Failure		401			{object}	utils.Response								"Unauthorized"
// @Failure		404			{object}	utils.Response								"Facility not found"
// @Failure		500			{object}	utils.Response								"Internal server error"
// @Router			/facilities/{facility_id} [put]
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

// @Tags			Facility
// @Summary		Delete facility
// @Description	Delete a facility type
// @Produce		json
// @Security		BearerAuth
// @Param			facility_id	path		int					true	"Facility ID"
// @Success		200			{object}	map[string]string	"Facility deleted successfully"
// @Failure		401			{object}	utils.Response		"Unauthorized"
// @Failure		404			{object}	utils.Response		"Facility not found"
// @Failure		500			{object}	utils.Response		"Internal server error"
// @Router			/facilities/{facility_id} [delete]
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
//
//	@Tags			Room Facility
//	@Summary		Add facility to room
//	@Description	Add a facility to a specific room
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateRoomFacilityRequest					true	"Create Room Facility Request"
//	@Success		201		{object}	utils.Response{data=dto.RoomFacilityResponse}	"Room facility created successfully"
//	@Failure		400		{object}	utils.Response									"Invalid request"
//	@Failure		401		{object}	utils.Response									"Unauthorized"
//	@Failure		404		{object}	utils.Response									"Room not found"
//	@Failure		500		{object}	utils.Response									"Internal server error"
//	@Router			/room-facilities/ [post]
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

// @Tags			Room Facility
// @Summary		Remove facility from room
// @Description	Remove a facility from a specific room
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.DeleteRoomFacilityRequest	true	"Delete Room Facility Request"
// @Success		200		{object}	map[string]string				"Room facility deleted successfully"
// @Failure		400		{object}	utils.Response					"Invalid request"
// @Failure		401		{object}	utils.Response					"Unauthorized"
// @Failure		404		{object}	utils.Response					"Room or facility not found"
// @Failure		500		{object}	utils.Response					"Internal server error"
// @Router			/room-facilities/ [delete]
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
