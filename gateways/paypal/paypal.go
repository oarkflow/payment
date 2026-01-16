package paypal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for PayPal
type Gateway struct {
	config *payment.GatewayConfig
}

// New creates a new PayPal gateway instance
func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://api.sandbox.paypal.com"
		} else {
			config.BaseURL = "https://api.paypal.com"
		}
	}
	if config.Currency == "" {
		config.Currency = "USD"
	}
	return &Gateway{config: config}
}

func (p *Gateway) GetName() string   { return "PayPal" }
func (p *Gateway) GetMethod() string { return "paypal" }

// InitiatePayment initiates a payment through PayPal
func (p *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// In a real implementation, this would call PayPal's Orders API
	orderID := fmt.Sprintf("PAYPAL-%d", time.Now().UnixNano())
	paymentURL := fmt.Sprintf("%s/checkoutnow?token=%s", p.config.BaseURL, orderID)

	return &payment.PaymentResponse{
		Success:       true,
		PaymentURL:    paymentURL,
		TransactionID: orderID,
		OrderID:       req.OrderID,
		Message:       "PayPal order created successfully",
	}, nil
}

// VerifyPayment verifies a payment with PayPal
func (p *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	// In a real implementation, this would call PayPal's Orders API to capture the payment
	return &payment.VerificationResponse{
		Success:       true,
		Status:        payment.StatusCompleted,
		TransactionID: req.TransactionID,
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		PaidAmount:    req.Amount,
		Message:       "Payment captured successfully",
	}, nil
}

// RefundPayment processes a refund through PayPal
func (p *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// In a real implementation, this would call PayPal's refund API
	return &payment.RefundResponse{
		Success:  true,
		RefundID: fmt.Sprintf("REF-%d", time.Now().UnixNano()),
		Message:  "Refund processed successfully",
	}, nil
}

// GetStatus retrieves the status of a payment from PayPal
func (p *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	// In a real implementation, this would call PayPal's Orders API to get order details
	amount:= money.New(0, money.MustCurrency(p.config.Currency))
	return &payment.StatusResponse{
		Status:        payment.StatusCompleted,
		TransactionID: txnID,
		Amount:        amount,
	}, nil
}
