package models

import (
	"time"
)

type Payment struct {
	ID            uint               `json:"id" gorm:"primaryKey" `
	BookingID     uint               `json:"booking_id" gorm:"index;not null" `
	ExternalID    string             `json:"external_id" gorm:"uniqueIndex;not null" `
	PaymentMethod *PaymentMethodType `json:"payment_method" gorm:"type:payment_method_type" `
	Amount        float64            `json:"amount" gorm:"type:decimal(12,2);not null" `
	Status        PaymentStatus      `json:"status" gorm:"type:payment_status;default:'pending'" `
	CheckoutURL   *string            `json:"checkout_url"`
	PaidAt        *time.Time         `json:"paid_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`

	Booking Booking `json:"-" gorm:"foreignKey:BookingID" `
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
	MethodEWallet      PaymentMethodType = "ewallet"
	MethodQRIS         PaymentMethodType = "qris"
	MethodRetail       PaymentMethodType = "retail_outlet"
)
