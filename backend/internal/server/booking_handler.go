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

// @Tags			Booking
// @Summary		Create a new booking
// @Description	Create a new room booking
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.CreateBookingRequest					true	"Create Booking Request"
// @Success		201		{object}	utils.Response{data=dto.BookingResponse}	"Booking created successfully"
// @Failure		400		{object}	utils.Response								"Invalid request or no available rooms"
// @Failure		401		{object}	utils.Response								"Unauthorized"
// @Failure		404		{object}	utils.Response								"Room not found"
// @Failure		500		{object}	utils.Response								"Internal server error"
// @Router			/booking/ [post]
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
