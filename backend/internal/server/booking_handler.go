package server

import (
	"errors"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService services.BookingService
}

func NewBookingHandler(bookingService services.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	booking := api.Group("/booking")
	booking.Use(middlewares...)
	booking.POST("/", h.CreateBooking)
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	userID := c.GetUint("user_id")
	booking, err := h.bookingService.CreateBooking(userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRoomNotFound):
			utils.NotFoundResponse(c, "room not found", nil)
		case errors.Is(err, services.ErrNoRoomsAvailable):
			utils.BadRequestResponse(c, "no available rooms", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to create booking", err)
		}
		return
	}

	utils.CreatedResponse(c, "booking created successfully", booking)
}
