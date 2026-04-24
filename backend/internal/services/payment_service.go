package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrBookingNotOwned      = errors.New("booking not owned by user")
	ErrBookingAlreadyPaid   = errors.New("booking already paid")
	ErrPaymentAlreadyExists = errors.New("payment already exists for this booking")
	ErrMidtransFailed       = errors.New("failed to create midtrans transaction")
)

type PaymentService interface {
	CreatePayment(userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	HandleWebhook(req *dto.MidtransWebhookRequest) error
}

type paymentService struct {
	db     *gorm.DB
	config *config.MidtransConfig
	snap   snap.Client
	core   coreapi.Client
}

func NewPaymentService(db *gorm.DB, cfg *config.MidtransConfig) PaymentService {
	var env midtrans.EnvironmentType
	if cfg.IsProduction {
		env = midtrans.Production
	} else {
		env = midtrans.Sandbox
	}

	snapClient := snap.Client{}
	snapClient.New(cfg.ServerKey, env)

	coreClient := coreapi.Client{}
	coreClient.New(cfg.ServerKey, env)

	return &paymentService{
		db:     db,
		config: cfg,
		snap:   snapClient,
		core:   coreClient,
	}
}

func (s *paymentService) CreatePayment(userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	var payment models.Payment

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var booking models.Booking
		if err := tx.Preload("User").Preload("Room").First(&booking, req.BookingID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBookingNotFound
			}
			return fmt.Errorf("failed to get booking: %w", err)
		}

		if booking.UserID != userID {
			return ErrBookingNotOwned
		}

		if booking.Status == models.BookingPaid {
			return ErrBookingAlreadyPaid
		}

		var existingPayment models.Payment
		if err := tx.Where("booking_id = ?", req.BookingID).First(&existingPayment).Error; err == nil {
			return ErrPaymentAlreadyExists
		}

		if booking.TotalPrice == nil || *booking.TotalPrice == 0 {
			totalPrice := math.Round(booking.Room.PricePerMonth.InexactFloat64() * float64(booking.DurationsMonths))
			booking.TotalPrice = &totalPrice
			if err := tx.Model(&booking).Update("total_price", totalPrice).Error; err != nil {
				return fmt.Errorf("failed to update booking total price: %w", err)
			}
		}

		orderID := fmt.Sprintf("INV-%s", uuid.New().String())
		snapReq := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: int64(math.Round(*booking.TotalPrice)),
			},
			CustomerDetail: &midtrans.CustomerDetails{
				FName: booking.User.Name,
				Email: booking.User.Email,
			},
		}

		snapResp, err := s.snap.CreateTransaction(snapReq)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMidtransFailed, err)
		}

		payment = models.Payment{
			BookingID:     req.BookingID,
			ExternalID:    orderID,
			InvoiceNumber: &orderID,
			Amount:        *booking.TotalPrice,
			Status:        models.PaymentPending,
			SnapToken:     &snapResp.Token,
			CheckoutURL:   &snapResp.RedirectURL,
		}

		if err := tx.Create(&payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return toPaymentResponse(&payment), nil
}

func (s *paymentService) HandleWebhook(req *dto.MidtransWebhookRequest) error {
	statusResp, err := s.core.CheckTransaction(req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to verify transaction status: %w", err)
	}

	var payment models.Payment
	if err := s.db.Where("external_id = ?", req.OrderID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("failed to find payment: %w", err)
	}

	status := mapMidtransStatus(statusResp.TransactionStatus, statusResp.FraudStatus)

	updates := map[string]interface{}{
		"status":            status,
		"transaction_id":    statusResp.TransactionID,
		"payment_type":      statusResp.PaymentType,
		"midtrans_response": json.RawMessage(mustMarshal(req)),
	}

	if len(statusResp.VaNumbers) > 0 {
		updates["va_number"] = fmt.Sprintf("%s-%s", statusResp.VaNumbers[0].Bank, statusResp.VaNumbers[0].VANumber)
	}

	if statusResp.ExpiryTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", statusResp.ExpiryTime); err == nil {
			updates["expiry_time"] = t
		}
	}

	if status == models.PaymentSuccess && payment.PaidAt == nil {
		now := time.Now()
		updates["paid_at"] = now
	}

	if err := s.db.Model(&payment).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if status == models.PaymentSuccess {
		if err := s.db.Model(&models.Booking{}).Where("id = ?", payment.BookingID).Update("status", models.BookingPaid).Error; err != nil {
			return fmt.Errorf("failed to update booking status: %w", err)
		}
	}

	return nil
}

func mapMidtransStatus(transactionStatus, fraudStatus string) models.PaymentStatus {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			return models.PaymentPending
		}
		if fraudStatus == "accept" {
			return models.PaymentSuccess
		}
		return models.PaymentSuccess
	case "settlement":
		return models.PaymentSuccess
	case "deny", "cancel", "failure":
		return models.PaymentFailed
	case "expire":
		return models.PaymentExpired
	case "pending":
		return models.PaymentPending
	default:
		return models.PaymentPending
	}
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func toPaymentResponse(p *models.Payment) *dto.PaymentResponse {
	resp := &dto.PaymentResponse{
		ID:            p.ID,
		BookingID:     p.BookingID,
		ExternalID:    p.ExternalID,
		InvoiceNumber: p.InvoiceNumber,
		Amount:        p.Amount,
		Status:        string(p.Status),
		SnapToken:     p.SnapToken,
		CheckoutURL:   p.CheckoutURL,
		TransactionID: p.TransactionID,
		PaymentType:   p.PaymentType,
		VANumber:      p.VANumber,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
	}

	if p.PaidAt != nil {
		formatted := p.PaidAt.Format(time.RFC3339)
		resp.PaidAt = &formatted
	}

	return resp
}
