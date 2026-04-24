package models

import (
	"encoding/json"
	"time"
)

type Payment struct {
	ID               uint               `json:"id" gorm:"primaryKey"`
	BookingID        uint               `json:"booking_id" gorm:"index;not null"`
	ExternalID       string             `json:"external_id" gorm:"uniqueIndex;not null"`
	InvoiceNumber    *string            `json:"invoice_number" gorm:"uniqueIndex"`
	PaymentMethod    *PaymentMethodType `json:"payment_method" gorm:"type:payment_method_type"`
	Amount           float64            `json:"amount" gorm:"type:decimal(12,2);not null"`
	Status           PaymentStatus      `json:"status" gorm:"type:payment_status;default:'pending'"`
	SnapToken        *string            `json:"snap_token"`
	CheckoutURL      *string            `json:"checkout_url"`
	TransactionID    *string            `json:"transaction_id" gorm:"uniqueIndex"`
	PaymentType      *string            `json:"payment_type"`
	VANumber         *string            `json:"va_number"`
	ExpiryTime       *time.Time         `json:"expiry_time"`
	MidtransResponse json.RawMessage    `json:"midtrans_response" gorm:"type:jsonb"`
	PaidAt           *time.Time         `json:"paid_at"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`

	Booking Booking `json:"-" gorm:"foreignKey:BookingID"`
}

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentSuccess PaymentStatus = "success"
	PaymentFailed  PaymentStatus = "failed"
	PaymentExpired PaymentStatus = "expired"
)

type PaymentMethodType string

const (
	MethodBankTransfer PaymentMethodType = "bank_transfer"
	MethodCreditCard   PaymentMethodType = "credit_card"
	MethodGoPay        PaymentMethodType = "gopay"
	MethodShopeePay    PaymentMethodType = "shopeepay"
	MethodQRIS         PaymentMethodType = "qris"
	MethodCStore       PaymentMethodType = "cstore"
)
