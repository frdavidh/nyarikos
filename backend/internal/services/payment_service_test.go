package services

import (
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapMidtransStatus(t *testing.T) {
	tests := []struct {
		name              string
		transactionStatus string
		fraudStatus       string
		expected          models.PaymentStatus
	}{
		{"settlement success", "settlement", "", models.PaymentSuccess},
		{"capture with accept fraud", "capture", "accept", models.PaymentSuccess},
		{"capture with challenge fraud", "capture", "challenge", models.PaymentPending},
		{"capture with empty fraud", "capture", "", models.PaymentSuccess},
		{"pending status", "pending", "", models.PaymentPending},
		{"deny status", "deny", "", models.PaymentFailed},
		{"cancel status", "cancel", "", models.PaymentFailed},
		{"failure status", "failure", "", models.PaymentFailed},
		{"expire status", "expire", "", models.PaymentExpired},
		{"unknown status", "unknown", "", models.PaymentPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMidtransStatus(tt.transactionStatus, tt.fraudStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPaymentResponse_WithAllFields(t *testing.T) {
	invoice := "INV-001"
	snapToken := "token-123"
	checkoutURL := "https://checkout.url"
	txID := "tx-123"
	paymentType := "gopay"
	vaNumber := "bca-123456"
	now := time.Now()

	payment := &models.Payment{
		ID:            1,
		BookingID:     2,
		ExternalID:    "INV-001",
		InvoiceNumber: &invoice,
		Amount:        1500000,
		Status:        models.PaymentPending,
		SnapToken:     &snapToken,
		CheckoutURL:   &checkoutURL,
		TransactionID: &txID,
		PaymentType:   &paymentType,
		VANumber:      &vaNumber,
		PaidAt:        &now,
		CreatedAt:     now,
	}

	resp := toPaymentResponse(payment)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, uint(2), resp.BookingID)
	assert.Equal(t, "INV-001", resp.ExternalID)
	assert.Equal(t, &invoice, resp.InvoiceNumber)
	assert.Equal(t, 1500000.0, resp.Amount)
	assert.Equal(t, "pending", resp.Status)
	assert.Equal(t, &snapToken, resp.SnapToken)
	assert.Equal(t, &checkoutURL, resp.CheckoutURL)
	assert.Equal(t, &txID, resp.TransactionID)
	assert.Equal(t, &paymentType, resp.PaymentType)
	assert.Equal(t, &vaNumber, resp.VANumber)
	assert.Equal(t, now.Format(time.RFC3339), *resp.PaidAt)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestToPaymentResponse_WithoutOptionalFields(t *testing.T) {
	now := time.Now()
	payment := &models.Payment{
		ID:         1,
		BookingID:  2,
		ExternalID: "INV-002",
		Amount:     2000000,
		Status:     models.PaymentPending,
		CreatedAt:  now,
	}

	resp := toPaymentResponse(payment)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, uint(2), resp.BookingID)
	assert.Equal(t, "INV-002", resp.ExternalID)
	assert.Nil(t, resp.InvoiceNumber)
	assert.Equal(t, 2000000.0, resp.Amount)
	assert.Equal(t, "pending", resp.Status)
	assert.Nil(t, resp.SnapToken)
	assert.Nil(t, resp.CheckoutURL)
	assert.Nil(t, resp.TransactionID)
	assert.Nil(t, resp.PaymentType)
	assert.Nil(t, resp.VANumber)
	assert.Nil(t, resp.PaidAt)
}

func TestCreatePayment_BookingNotFound(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	service := NewPaymentService(db, cfg)

	req := &dto.CreatePaymentRequest{BookingID: 999}
	_, err := service.CreatePayment(1, req)

	assert.ErrorIs(t, err, ErrBookingNotFound)
}

func TestCreatePayment_BookingNotOwned(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	service := NewPaymentService(db, cfg)

	user := models.User{Name: "Test", Email: "test@test.com"}
	require.NoError(t, db.Create(&user).Error)

	owner := models.User{Name: "Owner", Email: "owner@test.com"}
	require.NoError(t, db.Create(&owner).Error)

	room := models.Room{KostID: 1, RoomType: "Standard", PricePerMonth: decimal.NewFromInt(1000000), TotalRooms: 2}
	require.NoError(t, db.Create(&room).Error)

	booking := models.Booking{
		UserID:          owner.ID,
		RoomID:          room.ID,
		StartDate:       time.Now(),
		EndDate:         time.Now().AddDate(0, 2, 0),
		DurationsMonths: 2,
	}
	require.NoError(t, db.Create(&booking).Error)

	req := &dto.CreatePaymentRequest{BookingID: booking.ID}
	_, err := service.CreatePayment(user.ID, req)

	assert.ErrorIs(t, err, ErrBookingNotOwned)
}

func TestCreatePayment_BookingAlreadyPaid(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	service := NewPaymentService(db, cfg)

	user := models.User{Name: "Test", Email: "test@test.com"}
	require.NoError(t, db.Create(&user).Error)

	room := models.Room{KostID: 1, RoomType: "Standard", PricePerMonth: decimal.NewFromInt(1000000), TotalRooms: 2}
	require.NoError(t, db.Create(&room).Error)

	booking := models.Booking{
		UserID:          user.ID,
		RoomID:          room.ID,
		StartDate:       time.Now(),
		EndDate:         time.Now().AddDate(0, 2, 0),
		DurationsMonths: 2,
		Status:          models.BookingPaid,
	}
	require.NoError(t, db.Create(&booking).Error)

	req := &dto.CreatePaymentRequest{BookingID: booking.ID}
	_, err := service.CreatePayment(user.ID, req)

	assert.ErrorIs(t, err, ErrBookingAlreadyPaid)
}

func TestCreatePayment_PaymentAlreadyExists(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	service := NewPaymentService(db, cfg)

	user := models.User{Name: "Test", Email: "test@test.com"}
	require.NoError(t, db.Create(&user).Error)

	room := models.Room{KostID: 1, RoomType: "Standard", PricePerMonth: decimal.NewFromInt(1000000), TotalRooms: 2}
	require.NoError(t, db.Create(&room).Error)

	booking := models.Booking{
		UserID:          user.ID,
		RoomID:          room.ID,
		StartDate:       time.Now(),
		EndDate:         time.Now().AddDate(0, 2, 0),
		DurationsMonths: 2,
		Status:          models.BookingPending,
	}
	require.NoError(t, db.Create(&booking).Error)

	existing := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-OLD",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&existing).Error)

	req := &dto.CreatePaymentRequest{BookingID: booking.ID}
	_, err := service.CreatePayment(user.ID, req)

	assert.ErrorIs(t, err, ErrPaymentAlreadyExists)
}
