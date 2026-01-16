package razorpay

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for Razorpay
type Gateway struct {
	config *payment.GatewayConfig
}

// New creates a new Razorpay gateway instance
func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.razorpay.com"
	}
	if config.Currency == "" {
		config.Currency = "INR"
	}
	return &Gateway{config: config}
}

func (r *Gateway) GetName() string   { return "Razorpay" }
func (r *Gateway) GetMethod() string { return "razorpay" }

// InitiatePayment initiates a payment through Razorpay
func (r *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// In a real implementation, this would call Razorpay's Orders API
	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	paymentURL := fmt.Sprintf("%s/checkout/%s", r.config.BaseURL, orderID)

	return &payment.PaymentResponse{
		Success:       true,
		PaymentURL:    paymentURL,
		TransactionID: orderID,
		OrderID:       req.OrderID,
		Message:       "Order created successfully",
	}, nil
}

// VerifyPayment verifies a payment with Razorpay
func (r *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	// In a real implementation, this would verify the signature and call Razorpay's API
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

// RefundPayment processes a refund through Razorpay
func (r *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// In a real implementation, this would call Razorpay's refund API
	return &payment.RefundResponse{
		Success:  true,
		RefundID: fmt.Sprintf("rfnd_%d", time.Now().UnixNano()),
		Message:  "Refund processed successfully",
	}, nil
}

// GetStatus retrieves the status of a payment from Razorpay
func (r *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	// In a real implementation, this would call Razorpay's API
	amount:= money.New(0, money.MustCurrency(r.config.Currency))
	return &payment.StatusResponse{
		Status:        payment.StatusCompleted,
		TransactionID: txnID,
		Amount:        amount,
	}, nil
}
