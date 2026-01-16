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
	registry  *GatewayRegistry
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
		registry:  NewGatewayRegistry(),
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

// SetRegistry sets a custom gateway registry
func (pm *PaymentManager) SetRegistry(registry *GatewayRegistry) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.registry = registry
}

// GetRegistry returns the gateway registry
func (pm *PaymentManager) GetRegistry() *GatewayRegistry {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry
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

// GetAvailableGatewaysForCountry returns all available and configured gateways for a country
func (pm *PaymentManager) GetAvailableGatewaysForCountry(country Country) []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Get all gateways that are available in the registry for this country
	availableInRegistry := pm.registry.GetAvailableGateways(country)

	// Filter to only include gateways that are actually configured
	configured := []string{}
	for _, method := range availableInRegistry {
		if _, ok := pm.gateways[method]; ok {
			configured = append(configured, method)
		}
	}

	return configured
}

// GetRecommendedGateway returns the highest priority gateway for a country
func (pm *PaymentManager) GetRecommendedGateway(country Country) (string, error) {
	available := pm.GetAvailableGatewaysForCountry(country)
	if len(available) == 0 {
		return "", fmt.Errorf("no gateways available for country %s", country)
	}
	return available[0], nil
}

// InitiatePaymentForCountry initiates payment using the best gateway for a country
func (pm *PaymentManager) InitiatePaymentForCountry(ctx context.Context, country Country, req *PaymentRequest) (*PaymentResponse, error) {
	method, err := pm.GetRecommendedGateway(country)
	if err != nil {
		return nil, err
	}
	return pm.InitiatePayment(ctx, method, req)
}

// InitiatePaymentWithMethod initiates payment with validation for country
func (pm *PaymentManager) InitiatePaymentWithMethod(ctx context.Context, country Country, method string, req *PaymentRequest) (*PaymentResponse, error) {
	// Validate that the gateway is available for this country
	if err := pm.registry.ValidateGatewayForCountry(country, method); err != nil {
		return nil, err
	}

	// Check if gateway is configured
	if _, err := pm.GetGateway(method); err != nil {
		return nil, fmt.Errorf("gateway %s is available but not configured: %w", method, err)
	}

	return pm.InitiatePayment(ctx, method, req)
}

// GetGatewayRecommendations returns detailed recommendations for a country
func (pm *PaymentManager) GetGatewayRecommendations(country Country) []GatewayRecommendation {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	recommendations := pm.registry.GetRecommendations(country)

	// Update availability based on what's actually configured
	for i := range recommendations {
		_, configured := pm.gateways[recommendations[i].Method]
		recommendations[i].Available = configured
	}

	return recommendations
}

// ValidateGatewayForCountry checks if a gateway is both available and configured for a country
func (pm *PaymentManager) ValidateGatewayForCountry(country Country, method string) error {
	// Check registry
	if err := pm.registry.ValidateGatewayForCountry(country, method); err != nil {
		return err
	}

	// Check if configured
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if _, ok := pm.gateways[method]; !ok {
		return fmt.Errorf("gateway %s is not configured", method)
	}

	return nil
}

// IsGatewayAvailable checks if a gateway is available for a country
// Returns true if the gateway is registered in the registry for that country
func (pm *PaymentManager) IsGatewayAvailable(country Country, method string) bool {
	return pm.registry.IsGatewayAvailable(country, method)
}
