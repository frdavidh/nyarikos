package server

import (
	"net/http"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestPaymentHandler_CreatePayment_Success(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(&dto.PaymentResponse{
		ID:          1,
		BookingID:   1,
		ExternalID:  "INV-test-123",
		Amount:      1500000,
		Status:      "pending",
		SnapToken:   strPtr("snap-token-123"),
		CheckoutURL: strPtr("https://app.sandbox.midtrans.com/snap/v2/vtweb/token-123"),
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "payment created successfully", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	w := makeRequest(t, router, "POST", "/api/v1/payments/", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp["success"].(bool))
}

func TestPaymentHandler_CreatePayment_BookingNotFound(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 999}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, services.ErrBookingNotFound)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "booking not found", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_BookingNotOwned(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, services.ErrBookingNotOwned)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "booking not owned by user", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_BookingAlreadyPaid(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, services.ErrBookingAlreadyPaid)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "booking already paid", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_PaymentAlreadyExists(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, services.ErrPaymentAlreadyExists)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "payment already exists for this booking", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_MidtransFailed(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, services.ErrMidtransFailed)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "failed to create midtrans transaction", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_CreatePayment_GenericError(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.CreatePaymentRequest{BookingID: 1}
	mockPayment.On("CreatePayment", uint(1), &reqBody).Return(nil, assert.AnError)

	w := makeRequest(t, router, "POST", "/api/v1/payments/", reqBody, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "failed to create payment", resp["message"])
	mockPayment.AssertExpectations(t)
}

func TestPaymentHandler_Webhook_Success(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-test-123",
	}

	mockPayment.On("HandleWebhook", &reqBody).Return(nil)

	w := makeRequest(t, router, "POST", "/api/v1/payments/webhook", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "webhook processed successfully", resp["message"])
}

func TestPaymentHandler_Webhook_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	w := makeRequest(t, router, "POST", "/api/v1/payments/webhook", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp["success"].(bool))
}

func TestPaymentHandler_Webhook_PaymentNotFound(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-nonexistent",
	}

	mockPayment.On("HandleWebhook", &reqBody).Return(services.ErrPaymentNotFound)

	w := makeRequest(t, router, "POST", "/api/v1/payments/webhook", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "payment not found", resp["message"])
}

func TestPaymentHandler_Webhook_GenericError(t *testing.T) {
	router := setupTestRouter()
	mockPayment := new(mockPaymentService)
	handler := NewPaymentHandler(mockPayment)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pencari"))

	reqBody := dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-test-123",
	}

	mockPayment.On("HandleWebhook", &reqBody).Return(assert.AnError)

	w := makeRequest(t, router, "POST", "/api/v1/payments/webhook", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}
