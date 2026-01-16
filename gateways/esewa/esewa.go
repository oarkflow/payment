package esewa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for eSewa
type Gateway struct {
	config *payment.GatewayConfig
	client *http.Client
}

func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://rc-epay.esewa.com.np"
		} else {
			config.BaseURL = "https://epay.esewa.com.np"
		}
	}
	if config.Currency == "" {
		config.Currency = "NPR"
	}
	return &Gateway{config: config, client: client}
}

func (e *Gateway) GetName() string   { return "eSewa" }
func (e *Gateway) GetMethod() string { return "esewa" }

func (e *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	params := url.Values{}
	amountStr := req.Amount.Format(money.WithLocale(money.LocaleNeNP), money.WithoutComma(), money.WithoutSymbol())
	params.Set("amt", amountStr)
	params.Set("psc", "0")
	params.Set("pdc", "0")
	params.Set("txAmt", "0")
	params.Set("tAmt", amountStr)
	params.Set("pid", req.OrderID)
	params.Set("scd", e.config.MerchantID)
	params.Set("su", req.SuccessURL)
	params.Set("fu", req.FailureURL)

	paymentURL := fmt.Sprintf("%s/api/epay/main/v2/form?%s", e.config.BaseURL, params.Encode())

	return &payment.PaymentResponse{
		Success:    true,
		PaymentURL: paymentURL,
		OrderID:    req.OrderID,
	}, nil
}

func (e *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	data := url.Values{}
	amountStr := req.Amount.Format(money.WithLocale(money.LocaleNeNP), money.WithoutComma(), money.WithoutSymbol())
	data.Set("amt", amountStr)
	data.Set("rid", req.RawData["refId"])
	data.Set("pid", req.OrderID)
	data.Set("scd", e.config.MerchantID)

	verifyURL := fmt.Sprintf("%s/api/epay/transaction/status/", e.config.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", verifyURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	status := payment.StatusFailed
	if result["status"] == "COMPLETE" {
		status = payment.StatusCompleted
	}

	return &payment.VerificationResponse{
		Success:       status == payment.StatusCompleted,
		Status:        status,
		TransactionID: req.RawData["refId"],
		OrderID:       req.OrderID,
		Amount:        req.Amount,
	}, nil
}

func (e *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, errors.New("refund not supported by eSewa API")
}

func (e *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	return nil, errors.New("status check requires order details")
}
