package dto

type CreatePaymentRequest struct {
	BookingID uint `json:"booking_id" binding:"required" example:"1"`
}

type PaymentResponse struct {
	ID            uint    `json:"id" example:"1"`
	BookingID     uint    `json:"booking_id" example:"1"`
	ExternalID    string  `json:"external_id" example:"INV-20240101-001"`
	InvoiceNumber *string `json:"invoice_number,omitempty"`
	Amount        float64 `json:"amount" example:"1500000.00"`
	Status        string  `json:"status" example:"pending"`
	SnapToken     *string `json:"snap_token,omitempty"`
	CheckoutURL   *string `json:"checkout_url,omitempty"`
	TransactionID *string `json:"transaction_id,omitempty"`
	PaymentType   *string `json:"payment_type,omitempty"`
	VANumber      *string `json:"va_number,omitempty"`
	PaidAt        *string `json:"paid_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

type MidtransWebhookRequest struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	ExpiryTime        string `json:"expiry_time,omitempty"`
}
