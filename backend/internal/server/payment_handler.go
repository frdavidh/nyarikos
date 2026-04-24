package server

import (
	"errors"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	payment := api.Group("/payments")
	payment.Use(middlewares...)
	payment.POST("/", h.CreatePayment)

	api.POST("/payments/webhook", h.Webhook)
}

// @Tags			Payment
// @Summary		Create a new payment
// @Description	Create a Midtrans Snap payment for a booking
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.CreatePaymentRequest					true	"Create Payment Request"
// @Success		201		{object}	utils.Response{data=dto.PaymentResponse}	"Payment created successfully"
// @Failure		400		{object}	utils.Response								"Invalid request"
// @Failure		401		{object}	utils.Response								"Unauthorized"
// @Failure		404		{object}	utils.Response								"Booking not found"
// @Failure		409		{object}	utils.Response								"Payment already exists or booking already paid"
// @Failure		500		{object}	utils.Response								"Internal server error"
// @Router			/payments/ [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	userID := c.GetUint("user_id")
	payment, err := h.paymentService.CreatePayment(userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBookingNotFound):
			utils.NotFoundResponse(c, "booking not found", nil)
		case errors.Is(err, services.ErrBookingNotOwned):
			utils.ForbiddenResponse(c, "booking not owned by user", nil)
		case errors.Is(err, services.ErrBookingAlreadyPaid):
			utils.BadRequestResponse(c, "booking already paid", nil)
		case errors.Is(err, services.ErrPaymentAlreadyExists):
			utils.BadRequestResponse(c, "payment already exists for this booking", nil)
		case errors.Is(err, services.ErrMidtransFailed):
			utils.InternalServerErrorResponse(c, "failed to create midtrans transaction", err)
		default:
			utils.InternalServerErrorResponse(c, "failed to create payment", err)
		}
		return
	}

	utils.CreatedResponse(c, "payment created successfully", payment)
}

// @Tags			Payment
// @Summary		Midtrans webhook
// @Description	Handle Midtrans payment notification webhook
// @Accept			json
// @Produce		json
// @Param			request	body		dto.MidtransWebhookRequest	true	"Midtrans Webhook Request"
// @Success		200		{object}	utils.Response				"Webhook processed successfully"
// @Failure		400		{object}	utils.Response				"Invalid request"
// @Failure		404		{object}	utils.Response				"Payment not found"
// @Failure		500		{object}	utils.Response				"Internal server error"
// @Router			/payments/webhook [post]
func (h *PaymentHandler) Webhook(c *gin.Context) {
	var req dto.MidtransWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	if err := h.paymentService.HandleWebhook(&req); err != nil {
		switch {
		case errors.Is(err, services.ErrPaymentNotFound):
			utils.NotFoundResponse(c, "payment not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to process webhook", err)
		}
		return
	}

	utils.SuccessResponse(c, "webhook processed successfully", nil)
}
