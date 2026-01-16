package payment

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Unified Payment Manager
type PaymentManager struct {
	gateways  map[string]Gateway
	factories map[string]GatewayFactory
	client    *http.Client
	mu        sync.RWMutex
}

func NewPaymentManager(timeout time.Duration) *PaymentManager {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	pm := &PaymentManager{
		gateways:  make(map[string]Gateway),
		factories: make(map[string]GatewayFactory),
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	// Note: Gateway factories should be registered via RegisterFactory()
	// before calling RegisterGatewayWithConfig()

	return pm
}

// RegisterFactory registers a gateway factory for dynamic gateway creation
func (pm *PaymentManager) RegisterFactory(method string, factory GatewayFactory) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.factories[method] = factory
}

// RegisterGateway registers a pre-configured gateway instance
func (pm *PaymentManager) RegisterGateway(method string, gateway Gateway) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.gateways[method] = gateway
}

// RegisterGatewayWithConfig creates and registers a gateway using its factory
func (pm *PaymentManager) RegisterGatewayWithConfig(method string, config *GatewayConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	factory, ok := pm.factories[method]
	if !ok {
		return fmt.Errorf("no factory registered for method: %s", method)
	}

	gateway := factory(config, pm.client)
	pm.gateways[method] = gateway
	return nil
}

func (pm *PaymentManager) GetGateway(method string) (Gateway, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	g, ok := pm.gateways[method]
	if !ok {
		return nil, fmt.Errorf("gateway %s not registered", method)
	}
	return g, nil
}

func (pm *PaymentManager) ListGateways() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	methods := make([]string, 0, len(pm.gateways))
	for method := range pm.gateways {
		methods = append(methods, method)
	}
	return methods
}

func (pm *PaymentManager) InitiatePayment(ctx context.Context, method string, req *PaymentRequest) (*PaymentResponse, error) {
	g, err := pm.GetGateway(method)
	if err != nil {
		return nil, err
	}
	return g.InitiatePayment(ctx, req)
}

func (pm *PaymentManager) VerifyPayment(ctx context.Context, method string, req *VerificationRequest) (*VerificationResponse, error) {
	g, err := pm.GetGateway(method)
	if err != nil {
		return nil, err
	}
	return g.VerifyPayment(ctx, req)
}

func (pm *PaymentManager) RefundPayment(ctx context.Context, method string, req *RefundRequest) (*RefundResponse, error) {
	g, err := pm.GetGateway(method)
	if err != nil {
		return nil, err
	}
	return g.RefundPayment(ctx, req)
}

func (pm *PaymentManager) GetStatus(ctx context.Context, method string, txnID string) (*StatusResponse, error) {
	g, err := pm.GetGateway(method)
	if err != nil {
		return nil, err
	}
	return g.GetStatus(ctx, txnID)
}
