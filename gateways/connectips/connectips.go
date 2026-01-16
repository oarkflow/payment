package connectips

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/oarkflow/money"
	"github.com/oarkflow/payment"
)

// Gateway implements payment.Gateway for ConnectIPS
type Gateway struct {
	config *payment.GatewayConfig
	client *http.Client
}

func New(config *payment.GatewayConfig, client *http.Client) payment.Gateway {
	if config.BaseURL == "" {
		if config.Sandbox {
			config.BaseURL = "https://uat.connectips.com:7443/connectipswebgw"
		} else {
			config.BaseURL = "https://www.connectips.com/connectipswebgw"
		}
	}
	if config.Currency == "" {
		config.Currency = "NPR"
	}
	return &Gateway{config: config, client: client}
}

func (c *Gateway) GetName() string   { return "ConnectIPS" }
func (c *Gateway) GetMethod() string { return "connectips" }

func (c *Gateway) generateHash(data string) string {
	h := hmac.New(sha512.New, []byte(c.config.SecretKey))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *Gateway) InitiatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	txnAmt := req.Amount.Format(money.WithLocale(money.LocaleNeNP), money.WithoutComma(), money.WithoutSymbol())

	hashData := fmt.Sprintf("%s,%s,%s", c.config.MerchantID, req.OrderID, txnAmt)
	signature := c.generateHash(hashData)

	payload := map[string]string{
		"MERCHANTID":  c.config.MerchantID,
		"APPID":       c.config.APIKey,
		"REFERENCEID": req.OrderID,
		"TXNAMT":      txnAmt,
		"REMARKS":     req.Description,
		"PARTICULARS": req.Description,
		"TOKEN":       signature,
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/ips/initiate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &payment.PaymentResponse{
		Success:       result["status"] == "success",
		PaymentURL:    result["url"].(string),
		TransactionID: result["token"].(string),
		OrderID:       req.OrderID,
	}, nil
}

func (c *Gateway) VerifyPayment(ctx context.Context, req *payment.VerificationRequest) (*payment.VerificationResponse, error) {
	hashData := fmt.Sprintf("%s,%s", c.config.MerchantID, req.TransactionID)
	signature := c.generateHash(hashData)

	payload := map[string]string{
		"MERCHANTID": c.config.MerchantID,
		"APPID":      c.config.APIKey,
		"TXNID":      req.TransactionID,
		"TOKEN":      signature,
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/ips/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	status := payment.StatusFailed
	if result["status"] == "SUCCESS" {
		status = payment.StatusCompleted
	}

	var amount money.Money
	if amt, ok := result["amount"].(string); ok {
		if floatAmt, err := strconv.ParseFloat(amt, 64); err == nil {
			amount = money.New(int64(floatAmt*100), money.MustCurrency(c.config.Currency))
		}
	}

	return &payment.VerificationResponse{
		Success:       status == payment.StatusCompleted,
		Status:        status,
		TransactionID: req.TransactionID,
		OrderID:       result["reference_id"].(string),
		Amount:        amount,
	}, nil
}

func (c *Gateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, errors.New("refund not implemented for ConnectIPS")
}

func (c *Gateway) GetStatus(ctx context.Context, txnID string) (*payment.StatusResponse, error) {
	vReq := &payment.VerificationRequest{TransactionID: txnID}
	vResp, err := c.VerifyPayment(context.Background(), vReq)
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
