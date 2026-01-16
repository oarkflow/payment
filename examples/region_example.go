package main

import (
	"fmt"

	"github.com/oarkflow/payment"
	"github.com/oarkflow/payment/setup"
)

func main() {
	// Example 1: Basic usage with region-based gateway selection
	example1BasicRegionSupport()

	// Example 2: Get available gateways for a country
	example2AvailableGateways()

	// Example 3: Gateway availability validation (important!)
	example3GatewayAvailability()

	// Example 4: Get gateway recommendations
	example4GatewayRecommendations()

	// Example 5: Multi-country payment processing
	example5MultiCountryPayments()
}

// Example 1: Basic setup with region support
func example1BasicRegionSupport() {
	fmt.Println("\n=== Example 1: Basic Region Support ===")
	fmt.Println("IMPORTANT: Not all payment gateways work in all countries!")
	fmt.Println("  ❌ Stripe, PayPal, Wise do NOT work in Nepal")
	fmt.Println("  ✅ eSewa, Khalti, IME Pay work in Nepal")

	// Setup payment manager with gateway configs
	configs := map[string]*payment.GatewayConfig{
		// Nepal gateways
		"esewa": {
			MerchantID: "EPAYTEST",
			SecretKey:  "8gBm/:&EnhH.1/q",
			Sandbox:    true,
		},
		"khalti": {
			APIKey:  "test_public_key_xxx",
			Sandbox: true,
		},
		// International gateways (don't work in Nepal!)
		"stripe": {
			APIKey:  "sk_test_xxx",
			Sandbox: true,
		},
		"paypal": {
			APIKey:  "paypal_client_id",
			Sandbox: true,
		},
	}

	// Use default registry with accurate country support
	pm := setup.SetupPaymentManagerWithDefaults(configs)

	// Check what gateways are available for Nepal
	// This will only return gateways that actually work in Nepal
	nepalGateways := pm.GetAvailableGatewaysForCountry(payment.CountryNepal)
	fmt.Printf("\n✅ Available gateways for Nepal: %v\n", nepalGateways)
	fmt.Println("   (Notice: Stripe and PayPal are NOT included)")

	// Check what gateways are available for USA
	usaGateways := pm.GetAvailableGatewaysForCountry(payment.CountryUSA)
	fmt.Printf("\n✅ Available gateways for USA: %v\n", usaGateways)

	// Verify Stripe doesn't work in Nepal
	if pm.IsGatewayAvailable(payment.CountryNepal, "stripe") {
		fmt.Println("\n❌ ERROR: Stripe should NOT be available in Nepal!")
	} else {
		fmt.Println("\n✅ Correct: Stripe is NOT available in Nepal")
	}

	// Verify eSewa only works in Nepal
	if pm.IsGatewayAvailable(payment.CountryUSA, "esewa") {
		fmt.Println("❌ ERROR: eSewa should NOT be available in USA!")
	} else {
		fmt.Println("✅ Correct: eSewa is NOT available in USA")
	}
}

// Example 2: Get available gateways for different countries
func example2AvailableGateways() {
	fmt.Println("\n=== Example 2: Available Gateways by Country ===")

	pm := setupTestPaymentManager()

	// Test various countries
	testCases := []struct {
		country  payment.Country
		expected string
	}{
		{payment.CountryNepal, "esewa, khalti, imepay, connectips (NO Stripe/PayPal)"},
		{payment.CountryIndia, "razorpay (NO Stripe for receiving)"},
		{payment.CountryUSA, "stripe, paypal"},
		{payment.CountryCanada, "stripe, paypal"},
		{payment.CountryUK, "stripe, paypal"},
	}

	for _, tc := range testCases {
		gateways := pm.GetAvailableGatewaysForCountry(tc.country)
		fmt.Printf("%-15s -> %v\n", tc.country, gateways)
		fmt.Printf("                   (Expected: %s)\n\n", tc.expected)
	}
}

// Example 3: Gateway availability validation - CRITICAL!
func example3GatewayAvailability() {
	fmt.Println("\n=== Example 3: Gateway Availability Validation ===")
	fmt.Println("Always check if a gateway works in the customer's country!")

	pm := setupTestPaymentManager()

	// Test cases showing what works and what doesn't
	testCases := []struct {
		country payment.Country
		gateway string
		works   bool
		reason  string
	}{
		{payment.CountryNepal, "esewa", true, "eSewa is Nepal's primary gateway"},
		{payment.CountryNepal, "khalti", true, "Khalti is popular in Nepal"},
		{payment.CountryNepal, "stripe", false, "Stripe does NOT support Nepal"},
		{payment.CountryNepal, "paypal", false, "PayPal does NOT support receiving in Nepal"},
		{payment.CountryIndia, "razorpay", true, "Razorpay is India-specific"},
		{payment.CountryIndia, "esewa", false, "eSewa is Nepal-only"},
		{payment.CountryUSA, "stripe", true, "Stripe works in USA"},
		{payment.CountryUSA, "paypal", true, "PayPal works in USA"},
		{payment.CountryUSA, "esewa", false, "eSewa is Nepal-only"},
	}

	for _, tc := range testCases {
		available := pm.IsGatewayAvailable(tc.country, tc.gateway)
		symbol := "✅"
		if !available {
			symbol = "❌"
		}

		match := "✓"
		if available != tc.works {
			match = "✗ MISMATCH!"
		}

		fmt.Printf("%s %s + %s = %v (%s) %s\n",
			symbol, tc.country, tc.gateway, available, tc.reason, match)
	}
}

// Example 4: Get detailed gateway recommendations
func example4GatewayRecommendations() {
	fmt.Println("\n=== Example 4: Gateway Recommendations ===")

	pm := setupTestPaymentManager()

	// Get recommendations for Nepal
	fmt.Println("\n📍 Recommendations for Nepal:")
	fmt.Println("   (Only gateways that actually work in Nepal)")
	recommendations := pm.GetGatewayRecommendations(payment.CountryNepal)

	for _, rec := range recommendations {
		status := "Not Configured"
		if rec.Available {
			status = "Configured ✓"
		}
		recommendTag := ""
		if rec.Recommended {
			recommendTag = " ⭐ RECOMMENDED"
		}
		fmt.Printf("  %d. %s (Scope: %s, Status: %s)%s\n",
			rec.Priority, rec.Method, rec.Scope, status, recommendTag)
	}

	// Get recommendations for USA
	fmt.Println("\n📍 Recommendations for USA:")
	recommendations = pm.GetGatewayRecommendations(payment.CountryUSA)

	for _, rec := range recommendations {
		status := "Not Configured"
		if rec.Available {
			status = "Configured ✓"
		}
		recommendTag := ""
		if rec.Recommended {
			recommendTag = " ⭐ RECOMMENDED"
		}
		fmt.Printf("  %d. %s (Scope: %s, Status: %s)%s\n",
			rec.Priority, rec.Method, rec.Scope, status, recommendTag)
	}
}

// Example 5: Multi-country payment processing
func example5MultiCountryPayments() {
	fmt.Println("\n=== Example 5: Multi-Country Payment Processing ===")
	fmt.Println("Showing how to handle payments from customers in different countries")

	pm := setupTestPaymentManager()

	// Simulate customers from different countries
	customers := []struct {
		name    string
		country payment.Country
		amount  float64
		currency string
	}{
		{"Ram Sharma", payment.CountryNepal, 1000, "NPR"},
		{"Raj Patel", payment.CountryIndia, 500, "INR"},
		{"John Smith", payment.CountryUSA, 100, "USD"},
		{"Jane Doe", payment.CountryCanada, 150, "CAD"},
	}

	for _, customer := range customers {
		fmt.Printf("\n👤 Customer: %s from %s\n", customer.name, customer.country)

		// Get available payment methods for this customer's country
		availableMethods := pm.GetAvailableGatewaysForCountry(customer.country)
		fmt.Printf("   Available payment methods: %v\n", availableMethods)

		// Get recommended gateway (highest priority)
		recommended, err := pm.GetRecommendedGateway(customer.country)
		if err != nil {
			fmt.Printf("   ❌ No gateway available: %v\n", err)
			continue
		}
		fmt.Printf("   ⭐ Recommended gateway: %s\n", recommended)

		// In real app, you would:
		// 1. Show only available methods in payment UI
		// 2. Use recommended gateway as default selection
		// 3. Process payment with selected gateway
		fmt.Printf("   💰 Processing %.2f %s via %s\n",
			customer.amount, customer.currency, recommended)
	}
}
// Helper function to setup a test payment manager
func setupTestPaymentManager() *payment.PaymentManager {
	configs := map[string]*payment.GatewayConfig{
		"esewa": {
			MerchantID: "EPAYTEST",
			SecretKey:  "8gBm/:&EnhH.1/q",
			Sandbox:    true,
		},
		"khalti": {
			APIKey:  "test_public_key_xxx",
			Sandbox: true,
		},
		"imepay": {
			MerchantID: "test_merchant",
			APIKey:     "test_key",
			Sandbox:    true,
		},
		"connectips": {
			MerchantID: "test_merchant",
			APIKey:     "test_key",
			Sandbox:    true,
		},
		"razorpay": {
			APIKey:    "rzp_test_xxx",
			SecretKey: "secret_xxx",
			Sandbox:   true,
		},
		"stripe": {
			APIKey:  "sk_test_xxx",
			Sandbox: true,
		},
		"paypal": {
			APIKey:     "paypal_client_id",
			SecretKey:  "paypal_secret",
			Sandbox:    true,
		},
	}

	// Use the default setup which includes accurate country registrations
	return setup.SetupPaymentManagerWithDefaults(configs)
}
