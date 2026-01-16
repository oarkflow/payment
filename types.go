package payment

import (
	"context"
	"net/http"
	"time"

	"github.com/oarkflow/money"
)

// Core types
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
	StatusCanceled  PaymentStatus = "canceled"
)

// Gateway interface - all payment providers must implement this
type Gateway interface {
	InitiatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	VerifyPayment(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error)
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
	GetStatus(ctx context.Context, txnID string) (*StatusResponse, error)
	GetName() string
	GetMethod() string
}

// WebhookHandler interface for handling payment callbacks
type WebhookHandler interface {
	ParseWebhook(req *http.Request) (*WebhookData, error)
	ValidateWebhook(req *http.Request) error
}

// Request/Response types
type PaymentRequest struct {
	Amount        money.Money       `json:"amount"`
	OrderID       string            `json:"order_id"`
	CustomerName  string            `json:"customer_name,omitempty"`
	CustomerEmail string            `json:"customer_email,omitempty"`
	CustomerPhone string            `json:"customer_phone,omitempty"`
	SuccessURL    string            `json:"success_url"`
	FailureURL    string            `json:"failure_url,omitempty"`
	ReturnURL     string            `json:"return_url,omitempty"`
	WebhookURL    string            `json:"webhook_url,omitempty"`
	Description   string            `json:"description,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type PaymentResponse struct {
	Success       bool              `json:"success"`
	PaymentURL    string            `json:"payment_url,omitempty"`
	TransactionID string            `json:"transaction_id,omitempty"`
	OrderID       string            `json:"order_id"`
	Message       string            `json:"message,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type VerificationRequest struct {
	TransactionID string            `json:"transaction_id,omitempty"`
	OrderID       string            `json:"order_id,omitempty"`
	Amount        money.Money       `json:"amount,omitempty"`
	RawData       map[string]string `json:"raw_data,omitempty"`
}

type VerificationResponse struct {
	Success       bool              `json:"success"`
	Status        PaymentStatus     `json:"status"`
	TransactionID string            `json:"transaction_id"`
	OrderID       string            `json:"order_id"`
	Amount        money.Money       `json:"amount"`
	PaidAmount    money.Money       `json:"paid_amount,omitempty"`
	Fee           money.Money       `json:"fee,omitempty"`
	Message       string            `json:"message,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type RefundRequest struct {
	TransactionID string      `json:"transaction_id"`
	Amount        money.Money `json:"amount"`
	Reason        string      `json:"reason,omitempty"`
}

type RefundResponse struct {
	Success  bool   `json:"success"`
	RefundID string `json:"refund_id,omitempty"`
	Message  string `json:"message,omitempty"`
}

type StatusResponse struct {
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id"`
	OrderID       string        `json:"order_id"`
	Amount        money.Money   `json:"amount"`
}

type WebhookData struct {
	TransactionID string            `json:"transaction_id"`
	OrderID       string            `json:"order_id"`
	Amount        money.Money       `json:"amount"`
	Status        PaymentStatus     `json:"status"`
	RawData       map[string]string `json:"raw_data"`
}

// Config for each gateway
type GatewayConfig struct {
	MerchantID  string
	SecretKey   string
	APIKey      string
	BaseURL     string
	Timeout     time.Duration
	Sandbox     bool
	Currency    string // Default currency for the gateway
	ExtraConfig map[string]interface{}
}

// GatewayFactory is a function that creates a gateway instance
type GatewayFactory func(config *GatewayConfig, client *http.Client) Gateway
