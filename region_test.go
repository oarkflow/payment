package payment

import (
	"testing"
)

func TestRegionMapping(t *testing.T) {
	tests := []struct {
		country Country
		want    Region
	}{
		{CountryNepal, RegionSouthAsia},
		{CountryIndia, RegionSouthAsia},
		{CountryUSA, RegionNorthAmerica},
		{CountrySingapore, RegionSoutheastAsia},
		{CountryUK, RegionEurope},
	}

	for _, tt := range tests {
		got := GetRegion(tt.country)
		if got != tt.want {
			t.Errorf("GetRegion(%s) = %s; want %s", tt.country, got, tt.want)
		}
	}
}

func TestGetCountriesInRegion(t *testing.T) {
	countries := GetCountriesInRegion(RegionSouthAsia)
	if len(countries) == 0 {
		t.Error("Expected countries in South Asia, got none")
	}

	// Check if Nepal is in South Asia
	found := false
	for _, c := range countries {
		if c == CountryNepal {
			found = true
			break
		}
	}
	if !found {
		t.Error("Nepal should be in South Asia region")
	}
}

func TestGatewayRegistry(t *testing.T) {
	registry := NewGatewayRegistry()

	// Register country-specific gateway
	registry.RegisterCountryGateway(CountryNepal, "esewa", 1)
	registry.RegisterCountryGateway(CountryNepal, "khalti", 2)

	// Register global gateway
	registry.RegisterGlobalGateway("stripe", 10)

	// Test availability
	if !registry.IsGatewayAvailable(CountryNepal, "esewa") {
		t.Error("ESewa should be available for Nepal")
	}

	if !registry.IsGatewayAvailable(CountryNepal, "stripe") {
		t.Error("Stripe (global) should be available for Nepal")
	}

	if registry.IsGatewayAvailable(CountryUSA, "esewa") {
		t.Error("ESewa should not be available for USA")
	}

	// Test get available gateways
	gateways := registry.GetAvailableGateways(CountryNepal)
	if len(gateways) != 3 {
		t.Errorf("Expected 3 gateways for Nepal, got %d", len(gateways))
	}

	// Check priority ordering (esewa should be first)
	if gateways[0] != "esewa" {
		t.Errorf("Expected esewa to be first (highest priority), got %s", gateways[0])
	}
}

func TestRegionGateway(t *testing.T) {
	registry := NewGatewayRegistry()

	// Register region-specific gateway
	registry.RegisterRegionGateway(RegionSouthAsia, "regional-pay", 5)

	// Should be available for all South Asian countries
	if !registry.IsGatewayAvailable(CountryNepal, "regional-pay") {
		t.Error("Regional gateway should be available for Nepal")
	}

	if !registry.IsGatewayAvailable(CountryIndia, "regional-pay") {
		t.Error("Regional gateway should be available for India")
	}

	if registry.IsGatewayAvailable(CountryUSA, "regional-pay") {
		t.Error("Regional gateway should not be available for USA")
	}
}

func TestGatewayPriority(t *testing.T) {
	registry := NewGatewayRegistry()

	registry.RegisterCountryGateway(CountryNepal, "low-priority", 10)
	registry.RegisterCountryGateway(CountryNepal, "high-priority", 1)
	registry.RegisterCountryGateway(CountryNepal, "mid-priority", 5)

	gateways := registry.GetAvailableGateways(CountryNepal)

	// Should be sorted by priority
	expected := []string{"high-priority", "mid-priority", "low-priority"}
	for i, want := range expected {
		if gateways[i] != want {
			t.Errorf("Position %d: got %s, want %s", i, gateways[i], want)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	// Test Nepal gateways
	if !registry.IsGatewayAvailable(CountryNepal, "esewa") {
		t.Error("ESewa should be in default registry for Nepal")
	}

	if !registry.IsGatewayAvailable(CountryNepal, "khalti") {
		t.Error("Khalti should be in default registry for Nepal")
	}

	// Test global gateways
	if !registry.IsGatewayAvailable(CountryUSA, "stripe") {
		t.Error("Stripe should be available globally")
	}

	if !registry.IsGatewayAvailable(CountryNepal, "stripe") {
		t.Error("Stripe should be available in Nepal")
	}

	// Test recommendations
	recs := registry.GetRecommendations(CountryNepal)
	if len(recs) == 0 {
		t.Error("Should have recommendations for Nepal")
	}

	// First recommendation should be country-specific
	if recs[0].Scope != "country" {
		t.Errorf("First recommendation should be country-specific, got %s", recs[0].Scope)
	}
}

func TestGatewayRecommendations(t *testing.T) {
	registry := NewGatewayRegistry()

	// Setup
	registry.RegisterCountryGateway(CountryNepal, "esewa", 1)
	registry.RegisterCountryGateway(CountryNepal, "khalti", 2)
	registry.RegisterRegionGateway(RegionSouthAsia, "regional", 5)
	registry.RegisterGlobalGateway("stripe", 10)

	recs := registry.GetRecommendations(CountryNepal)

	// Should have 4 recommendations
	if len(recs) != 4 {
		t.Errorf("Expected 4 recommendations, got %d", len(recs))
	}

	// Check scopes are correct
	scopes := map[string]int{"country": 0, "region": 0, "global": 0}
	for _, rec := range recs {
		scopes[rec.Scope]++
	}

	if scopes["country"] != 2 {
		t.Errorf("Expected 2 country-specific, got %d", scopes["country"])
	}
	if scopes["region"] != 1 {
		t.Errorf("Expected 1 region-specific, got %d", scopes["region"])
	}
	if scopes["global"] != 1 {
		t.Errorf("Expected 1 global, got %d", scopes["global"])
	}
}

func TestValidateGatewayForCountry(t *testing.T) {
	registry := NewGatewayRegistry()
	registry.RegisterCountryGateway(CountryNepal, "esewa", 1)
	registry.RegisterGlobalGateway("stripe", 10)

	// Valid cases
	if err := registry.ValidateGatewayForCountry(CountryNepal, "esewa"); err != nil {
		t.Errorf("Validation should pass for ESewa in Nepal: %v", err)
	}

	if err := registry.ValidateGatewayForCountry(CountryNepal, "stripe"); err != nil {
		t.Errorf("Validation should pass for Stripe in Nepal: %v", err)
	}

	// Invalid case
	if err := registry.ValidateGatewayForCountry(CountryUSA, "esewa"); err == nil {
		t.Error("Validation should fail for ESewa in USA")
	}
}
