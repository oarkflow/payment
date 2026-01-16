package imepay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for IMEPay
type Gateway struct {
	config *payment.GatewayConfig
	client *http.Client
}

func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://stg.imepay.com.np:7979/api/Web"
		} else {
			config.BaseURL = "https://payment.imepay.com.np:7979/api/Web"
		}
	}
	if config.Currency == "" {
		config.Currency = "NPR"
	}
	return &Gateway{config: config, client: client}
}

func (i *Gateway) GetName() string   { return "IMEPay" }
func (i *Gateway) GetMethod() string { return "imepay" }

func (i *Gateway) generateToken(data string) string {
	h := sha256.New()
	h.Write([]byte(data + i.config.SecretKey))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}

func (i *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	amount := req.Amount.Format(money.WithLocale(money.LocaleNeNP), money.WithoutComma(), money.WithoutSymbol())
	refID := req.OrderID

	tokenData := fmt.Sprintf("MerchantCode=%s,RefId=%s,TranAmount=%s", i.config.MerchantID, refID, amount)
	token := i.generateToken(tokenData)

	params := url.Values{}
	params.Set("MerchantCode", i.config.MerchantID)
	params.Set("RefId", refID)
	params.Set("TranAmount", amount)
	params.Set("Method", "GET")
	params.Set("ResponseUrl", req.SuccessURL)
	params.Set("CancelUrl", req.FailureURL)
	params.Set("TokenId", token)

	paymentURL := fmt.Sprintf("%s/Checkout?%s", i.config.BaseURL, params.Encode())

	return &payment.PaymentResponse{
		Success:    true,
		PaymentURL: paymentURL,
		OrderID:    refID,
	}, nil
}

func (i *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	msisdn := req.RawData["Msisdn"]
	refID := req.RawData["RefId"]
	txnID := req.RawData["TransactionId"]

	tokenData := fmt.Sprintf("Msisdn=%s,RefId=%s,TransactionId=%s", msisdn, refID, txnID)
	token := i.generateToken(tokenData)

	payload := map[string]string{
		"MerchantCode":  i.config.MerchantID,
		"RefId":         refID,
		"TransactionId": txnID,
		"Msisdn":        msisdn,
		"TokenId":       token,
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", i.config.BaseURL+"/Reconfirm", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	status := payment.StatusFailed
	if result["ResponseCode"] == "0" {
		status = payment.StatusCompleted
	}

	var amount money.Money
	if amt, ok := result["Amount"].(string); ok {
		if floatAmt, err := strconv.ParseFloat(amt, 64); err == nil {
			amount = money.New(int64(floatAmt*100), money.MustCurrency(i.config.Currency))
		}
	}

	return &payment.VerificationResponse{
		Success:       status == payment.StatusCompleted,
		Status:        status,
		TransactionID: txnID,
		OrderID:       refID,
		Amount:        amount,
	}, nil
}

func (i *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, errors.New("refund not implemented for IMEPay")
}

func (i *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	return nil, errors.New("status check requires additional data for IMEPay")
}
