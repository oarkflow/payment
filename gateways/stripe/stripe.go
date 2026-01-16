package stripe

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for Stripe
type Gateway struct {
	config *payment.GatewayConfig
	client *http.Client
}

// New creates a new Stripe gateway instance
func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://api.stripe.com/test"
		} else {
			config.BaseURL = "https://api.stripe.com"
		}
	}
	if config.Currency == "" {
		config.Currency = "USD"
	}
	return &Gateway{config: config, client: client}
}

func (s *Gateway) GetName() string   { return "Stripe" }
func (s *Gateway) GetMethod() string { return "stripe" }

// InitiatePayment initiates a payment through Stripe
func (s *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// In a real implementation, this would create a Stripe Checkout Session
	paymentURL := fmt.Sprintf("%s/checkout/%s", s.config.BaseURL, req.OrderID)

	return &payment.PaymentResponse{
		Success:       true,
		PaymentURL:    paymentURL,
		TransactionID: fmt.Sprintf("pi_%d", time.Now().UnixNano()),
		OrderID:       req.OrderID,
		Message:       "Payment session created successfully",
	}, nil
}

// VerifyPayment verifies a payment with Stripe
func (s *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	// In a real implementation, this would call Stripe's API to verify the payment
	return &payment.VerificationResponse{
		Success:       true,
		Status:        payment.StatusCompleted,
		TransactionID: req.TransactionID,
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		PaidAmount:    req.Amount,
		Message:       "Payment verified successfully",
	}, nil
}

// RefundPayment processes a refund through Stripe
func (s *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// In a real implementation, this would call Stripe's refund API
	return &payment.RefundResponse{
		Success:  true,
		RefundID: fmt.Sprintf("re_%d", time.Now().UnixNano()),
		Message:  "Refund processed successfully",
	}, nil
}

// GetStatus retrieves the status of a payment from Stripe
func (s *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	// In a real implementation, this would call Stripe's API to get payment status
	// For now, return a mock response
	amount:= money.New(0, money.MustCurrency(s.config.Currency))
	return &payment.StatusResponse{
		Status:        payment.StatusCompleted,
		TransactionID: txnID,
		Amount:        amount,
	}, nil
}
