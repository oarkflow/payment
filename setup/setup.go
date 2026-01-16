package setup

import (
	"log"
	"time"

	"github.com/oarkflow/payment"
	"github.com/oarkflow/payment/gateways/connectips"
	"github.com/oarkflow/payment/gateways/esewa"
	"github.com/oarkflow/payment/gateways/imepay"
	"github.com/oarkflow/payment/gateways/khalti"
	"github.com/oarkflow/payment/gateways/paypal"
	"github.com/oarkflow/payment/gateways/razorpay"
	"github.com/oarkflow/payment/gateways/stripe"
)

// SetupPaymentManager creates a fully configured payment manager with all gateways
func SetupPaymentManager(configs map[string]*payment.GatewayConfig) *payment.PaymentManager {
	pm := payment.NewPaymentManager(30 * time.Second)

	// Register built-in gateway factories - Nepal gateways
	pm.RegisterFactory("esewa", esewa.New)
	pm.RegisterFactory("khalti", khalti.New)
	pm.RegisterFactory("imepay", imepay.New)
	pm.RegisterFactory("connectips", connectips.New)

	// Register international gateway factories
	pm.RegisterFactory("stripe", stripe.New)
	pm.RegisterFactory("paypal", paypal.New)
	pm.RegisterFactory("razorpay", razorpay.New)

	// Register gateways with provided configs
	for method, config := range configs {
		if err := pm.RegisterGatewayWithConfig(method, config); err != nil {
			log.Printf("Error registering gateway %s: %v", method, err)
			continue
		}
	}

	return pm
}
// SetupPaymentManagerWithRegistry creates a payment manager with custom registry
func SetupPaymentManagerWithRegistry(
	configs map[string]*payment.GatewayConfig,
	registry *payment.GatewayRegistry,
) *payment.PaymentManager {
	pm := SetupPaymentManager(configs)
	pm.SetRegistry(registry)
	return pm
}

// createDefaultRegistry creates a registry with default country and region mappings
// Based on actual payment gateway support in each country
func createDefaultRegistry() *payment.GatewayRegistry {
	registry := payment.NewGatewayRegistry()

	// Register Nepal-specific payment gateways
	// Note: Stripe, PayPal, Wise do NOT support receiving payments in Nepal
	registry.RegisterCountryGateway(payment.CountryNepal, "esewa", 1)
	registry.RegisterCountryGateway(payment.CountryNepal, "khalti", 2)
	registry.RegisterCountryGateway(payment.CountryNepal, "imepay", 3)
	registry.RegisterCountryGateway(payment.CountryNepal, "connectips", 4)

	// Register India-specific payment gateways
	registry.RegisterCountryGateway(payment.CountryIndia, "razorpay", 1)
	registry.RegisterCountryGateway(payment.CountryIndia, "paytm", 2)

	// Register USA payment gateways
	registry.RegisterCountryGateway(payment.CountryUSA, "stripe", 1)
	registry.RegisterCountryGateway(payment.CountryUSA, "paypal", 2)

	// Register Canada payment gateways
	registry.RegisterCountryGateway(payment.CountryCanada, "stripe", 1)
	registry.RegisterCountryGateway(payment.CountryCanada, "paypal", 2)

	// Register UK payment gateways
	registry.RegisterCountryGateway(payment.CountryUK, "stripe", 1)
	registry.RegisterCountryGateway(payment.CountryUK, "paypal", 2)

	// Register by region - North America (US, Canada supported)
	registry.RegisterRegionGateway(payment.RegionNorthAmerica, "stripe", 1)
	registry.RegisterRegionGateway(payment.RegionNorthAmerica, "paypal", 2)

	// Register by region - Europe (most European countries supported)
	registry.RegisterRegionGateway(payment.RegionEurope, "stripe", 1)
	registry.RegisterRegionGateway(payment.RegionEurope, "paypal", 2)

	// Register by region - Oceania (Australia, New Zealand)
	registry.RegisterRegionGateway(payment.RegionOceania, "stripe", 1)
	registry.RegisterRegionGateway(payment.RegionOceania, "paypal", 2)

	return registry
}

// SetupPaymentManagerWithDefaults creates a payment manager with default registry
// This includes default Nepal gateway registrations
func SetupPaymentManagerWithDefaults(configs map[string]*payment.GatewayConfig) *payment.PaymentManager {
	pm := SetupPaymentManager(configs)
	registry := createDefaultRegistry()
	pm.SetRegistry(registry)
	return pm
}

// SetupForCountry creates a payment manager optimized for a specific country
func SetupForCountry(country payment.Country, configs map[string]*payment.GatewayConfig) *payment.PaymentManager {
	pm := SetupPaymentManager(configs)
	registry := createDefaultRegistry()
	pm.SetRegistry(registry)
	return pm
}

// SetupMultiRegion creates a payment manager for multiple regions
func SetupMultiRegion(configs map[string]*payment.GatewayConfig) *payment.PaymentManager {
	pm := SetupPaymentManager(configs)
	registry := createDefaultRegistry()
	pm.SetRegistry(registry)
	return pm
}
