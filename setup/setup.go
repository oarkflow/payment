package setup

import (
	"time"

	"github.com/oarkflow/payment"
	"github.com/oarkflow/payment/gateways/connectips"
	"github.com/oarkflow/payment/gateways/esewa"
	"github.com/oarkflow/payment/gateways/imepay"
	"github.com/oarkflow/payment/gateways/khalti"
)

// SetupPaymentManager creates a fully configured payment manager with all gateways
func SetupPaymentManager(configs map[string]*payment.GatewayConfig) *payment.PaymentManager {
	pm := payment.NewPaymentManager(30 * time.Second)

	// Register built-in gateway factories
	pm.RegisterFactory("esewa", esewa.New)
	pm.RegisterFactory("khalti", khalti.New)
	pm.RegisterFactory("imepay", imepay.New)
	pm.RegisterFactory("connectips", connectips.New)

	// Register gateways with provided configs
	for method, config := range configs {
		if err := pm.RegisterGatewayWithConfig(method, config); err != nil {
			// Log error or handle appropriately
			continue
		}
	}

	return pm
}
