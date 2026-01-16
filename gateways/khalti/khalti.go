package khalti

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for Khalti
type Gateway struct {
	config *payment.GatewayConfig
	client *http.Client
}

func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://a.khalti.com/api/v2"
		} else {
			config.BaseURL = "https://khalti.com/api/v2"
		}
	}
	if config.Currency == "" {
		config.Currency = "NPR"
	}
	return &Gateway{config: config, client: client}
}

func (k *Gateway) GetName() string   { return "Khalti" }
func (k *Gateway) GetMethod() string { return "khalti" }

func (k *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Khalti expects amount in paisa (1 NPR = 100 paisa)
	amountInPaisa := req.Amount.Amount()

	payload := map[string]interface{}{
		"return_url":          req.SuccessURL,
		"website_url":         req.ReturnURL,
		"amount":              amountInPaisa,
		"purchase_order_id":   req.OrderID,
		"purchase_order_name": req.Description,
		"customer_info": map[string]string{
			"name":  req.CustomerName,
			"email": req.CustomerEmail,
			"phone": req.CustomerPhone,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", k.config.BaseURL+"/epayment/initiate/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Key "+k.config.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("khalti error: %v", result)
	}

	return &payment.PaymentResponse{
		Success:       true,
		PaymentURL:    result["payment_url"].(string),
		TransactionID: result["pidx"].(string),
		OrderID:       req.OrderID,
	}, nil
}

func (k *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	payload := map[string]string{"pidx": req.TransactionID}
	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", k.config.BaseURL+"/epayment/lookup/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Key "+k.config.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	status := payment.StatusPending
	if result["status"] == payment.StatusCompleted {
		status = payment.StatusCompleted
	} else if result["status"] == payment.StatusPending {
		status = payment.StatusPending
	} else {
		status = payment.StatusFailed
	}

	var amount money.Money
	if amt, ok := result["total_amount"].(float64); ok {
		amount = money.New(int64(amt), money.MustCurrency(k.config.Currency))
	}

	var fee money.Money
	if feeAmt, ok := result["fee"].(float64); ok {
		fee = money.New(int64(feeAmt), money.MustCurrency(k.config.Currency))
	}

	return &payment.VerificationResponse{
		Success:       status == payment.StatusCompleted,
		Status:        status,
		TransactionID: req.TransactionID,
		OrderID:       result["purchase_order_id"].(string),
		Amount:        amount,
		Fee:           fee,
	}, nil
}

func (k *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, errors.New("refund not implemented for Khalti")
}

func (k *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	vReq := &payment.VerificationRequest{TransactionID: txnID}
	vResp, err := k.VerifyPayment(context.Background(), vReq)
	if err != nil {
		return nil, err
	}
	return &payment.StatusResponse{
		Status:        vResp.Status,
		TransactionID: vResp.TransactionID,
		OrderID:       vResp.OrderID,
		Amount:        vResp.Amount,
	}, nil
}
