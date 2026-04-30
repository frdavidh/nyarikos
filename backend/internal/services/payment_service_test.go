package services

import (
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockMidtransSnap struct {
	mock.Mock
}

func (m *mockMidtransSnap) CreateTransaction(req *snap.Request) (*snap.Response, *midtrans.Error) {
	args := m.Called(req)
	resp, _ := args.Get(0).(*snap.Response)
	err, _ := args.Get(1).(*midtrans.Error)
	if resp == nil {
		return nil, err
	}
	return resp, err
}

type mockMidtransCore struct {
	mock.Mock
}

func (m *mockMidtransCore) CheckTransaction(orderID string) (*coreapi.TransactionStatusResponse, *midtrans.Error) {
	args := m.Called(orderID)
	resp, _ := args.Get(0).(*coreapi.TransactionStatusResponse)
	err, _ := args.Get(1).(*midtrans.Error)
	if resp == nil {
		return nil, err
	}
	return resp, err
}

func newTestPaymentService(db *gorm.DB, cfg *config.MidtransConfig, snapClient midtransSnapClient, core midtransCoreClient) PaymentService {
	return &paymentService{
		db:     db,
		config: cfg,
		snap:   snapClient,
		core:   core,
	}
}

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

func TestCreatePayment_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	mockSnap.On("CreateTransaction", mock.AnythingOfType("*snap.Request")).Return(&snap.Response{
		Token:       "snap-token-123",
		RedirectURL: "https://checkout.url",
	}, (*midtrans.Error)(nil))

	req := &dto.CreatePaymentRequest{BookingID: booking.ID}
	resp, err := service.CreatePayment(user.ID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, booking.ID, resp.BookingID)
	assert.Equal(t, "snap-token-123", *resp.SnapToken)
	assert.Equal(t, "https://checkout.url", *resp.CheckoutURL)
	assert.Equal(t, "pending", resp.Status)
	mockSnap.AssertExpectations(t)
}

func TestCreatePayment_TotalPriceCalculated(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

	user := models.User{Name: "Test", Email: "test@test.com"}
	require.NoError(t, db.Create(&user).Error)

	room := models.Room{KostID: 1, RoomType: "Standard", PricePerMonth: decimal.NewFromInt(500000), TotalRooms: 2}
	require.NoError(t, db.Create(&room).Error)

	booking := models.Booking{
		UserID:          user.ID,
		RoomID:          room.ID,
		StartDate:       time.Now(),
		EndDate:         time.Now().AddDate(0, 3, 0),
		DurationsMonths: 3,
		Status:          models.BookingPending,
	}
	require.NoError(t, db.Create(&booking).Error)

	mockSnap.On("CreateTransaction", mock.AnythingOfType("*snap.Request")).Return(&snap.Response{
		Token:       "snap-token",
		RedirectURL: "https://checkout.url",
	}, (*midtrans.Error)(nil))

	req := &dto.CreatePaymentRequest{BookingID: booking.ID}
	resp, err := service.CreatePayment(user.ID, req)

	require.NoError(t, err)
	assert.Equal(t, 1500000.0, resp.Amount)

	var updated models.Booking
	err = db.First(&updated, booking.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updated.TotalPrice)
	assert.Equal(t, 1500000.0, *updated.TotalPrice)
}

func TestHandleWebhook_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-TEST-001",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-TEST-001").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "settlement",
		TransactionID:     "tx-123",
		PaymentType:       "bank_transfer",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-TEST-001",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)
	mockCore.AssertExpectations(t)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PaymentSuccess, updated.Status)
	assert.NotNil(t, updated.TransactionID)
	assert.Equal(t, "tx-123", *updated.TransactionID)
	assert.NotNil(t, updated.PaymentType)
	assert.Equal(t, "bank_transfer", *updated.PaymentType)
	assert.NotNil(t, updated.PaidAt)

	var updatedBooking models.Booking
	err = db.First(&updatedBooking, booking.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.BookingPaid, updatedBooking.Status)
}

func TestHandleWebhook_PaymentNotFound(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

	mockCore.On("CheckTransaction", "INV-MISSING").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "settlement",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-MISSING",
	}

	err := service.HandleWebhook(req)

	assert.ErrorIs(t, err, ErrPaymentNotFound)
}

func TestHandleWebhook_CheckTransactionError(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

	mockCore.On("CheckTransaction", "INV-ERROR").Return(nil, &midtrans.Error{
		Message:    "midtrans error",
		StatusCode: 500,
	})

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-ERROR",
	}

	err := service.HandleWebhook(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify transaction status")
}

func TestHandleWebhook_WithVANumber(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-VA-001",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-VA-001").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "pending",
		TransactionID:     "tx-va-123",
		PaymentType:       "bank_transfer",
		VaNumbers: []coreapi.VANumber{
			{Bank: "bca", VANumber: "1234567890"},
		},
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "pending",
		OrderID:           "INV-VA-001",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PaymentPending, updated.Status)
	assert.NotNil(t, updated.VANumber)
	assert.Equal(t, "bca-1234567890", *updated.VANumber)
}

func TestHandleWebhook_WithExpiryTime(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-EXP-001",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-EXP-001").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "pending",
		TransactionID:     "tx-exp-123",
		PaymentType:       "bank_transfer",
		ExpiryTime:        "2026-12-31 23:59:59",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "pending",
		OrderID:           "INV-EXP-001",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updated.ExpiryTime)
	expectedTime := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	assert.WithinDuration(t, expectedTime, *updated.ExpiryTime, time.Second)
}

func TestHandleWebhook_StatusFailed(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-FAIL-001",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-FAIL-001").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "deny",
		TransactionID:     "tx-deny-123",
		PaymentType:       "credit_card",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "deny",
		OrderID:           "INV-FAIL-001",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PaymentFailed, updated.Status)
	assert.Nil(t, updated.PaidAt)

	var updatedBooking models.Booking
	err = db.First(&updatedBooking, booking.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.BookingPending, updatedBooking.Status)
}

func TestHandleWebhook_StatusExpired(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-EXP-002",
		Amount:     2000000,
		Status:     models.PaymentPending,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-EXP-002").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "expire",
		TransactionID:     "tx-exp-123",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "expire",
		OrderID:           "INV-EXP-002",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PaymentExpired, updated.Status)
}

func TestHandleWebhook_PaidAtOnlySetOnce(t *testing.T) {
	db := setupTestDB(t)
	cfg := &testConfig().Midtrans
	mockSnap := new(mockMidtransSnap)
	mockCore := new(mockMidtransCore)

	service := newTestPaymentService(db, cfg, mockSnap, mockCore)

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

	firstPaidAt := time.Now().Add(-time.Hour)
	payment := models.Payment{
		BookingID:  booking.ID,
		ExternalID: "INV-PAID-ONCE",
		Amount:     2000000,
		Status:     models.PaymentSuccess,
		PaidAt:     &firstPaidAt,
	}
	require.NoError(t, db.Create(&payment).Error)

	mockCore.On("CheckTransaction", "INV-PAID-ONCE").Return(&coreapi.TransactionStatusResponse{
		TransactionStatus: "settlement",
		TransactionID:     "tx-123",
	}, (*midtrans.Error)(nil))

	req := &dto.MidtransWebhookRequest{
		TransactionStatus: "settlement",
		OrderID:           "INV-PAID-ONCE",
	}

	err := service.HandleWebhook(req)

	require.NoError(t, err)

	var updated models.Payment
	err = db.First(&updated, payment.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updated.PaidAt)
	assert.WithinDuration(t, firstPaidAt, *updated.PaidAt, time.Second)
}

var (
	_ midtransSnapClient = (*mockMidtransSnap)(nil)
	_ midtransCoreClient = (*mockMidtransCore)(nil)
)
