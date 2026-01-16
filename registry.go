package payment

import (
	"fmt"
	"sync"
)

// GatewayRegistry manages gateway availability by region and country
type GatewayRegistry struct {
	// Global gateways available everywhere
	globalGateways map[string]bool

	// Region-specific gateways
	regionGateways map[Region]map[string]bool

	// Country-specific gateways
	countryGateways map[Country]map[string]bool

	// Gateway priorities (lower number = higher priority)
	gatewayPriority map[string]int

	mu sync.RWMutex
}

// NewGatewayRegistry creates a new gateway registry
func NewGatewayRegistry() *GatewayRegistry {
	return &GatewayRegistry{
		globalGateways:  make(map[string]bool),
		regionGateways:  make(map[Region]map[string]bool),
		countryGateways: make(map[Country]map[string]bool),
		gatewayPriority: make(map[string]int),
	}
}

// RegisterGlobalGateway registers a gateway available globally
func (r *GatewayRegistry) RegisterGlobalGateway(method string, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.globalGateways[method] = true
	r.gatewayPriority[method] = priority
}

// RegisterRegionGateway registers a gateway for a specific region
func (r *GatewayRegistry) RegisterRegionGateway(region Region, method string, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.regionGateways[region] == nil {
		r.regionGateways[region] = make(map[string]bool)
	}
	r.regionGateways[region][method] = true
	r.gatewayPriority[method] = priority
}

// RegisterCountryGateway registers a gateway for a specific country
func (r *GatewayRegistry) RegisterCountryGateway(country Country, method string, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.countryGateways[country] == nil {
		r.countryGateways[country] = make(map[string]bool)
	}
	r.countryGateways[country][method] = true
	r.gatewayPriority[method] = priority
}

// GetAvailableGateways returns all available gateways for a country, sorted by priority
func (r *GatewayRegistry) GetAvailableGateways(country Country) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gatewaysMap := make(map[string]bool)

	// Add global gateways
	for method := range r.globalGateways {
		gatewaysMap[method] = true
	}

	// Add region gateways
	region := GetRegion(country)
	if regionGateways, ok := r.regionGateways[region]; ok {
		for method := range regionGateways {
			gatewaysMap[method] = true
		}
	}

	// Add country-specific gateways (highest priority)
	if countryGateways, ok := r.countryGateways[country]; ok {
		for method := range countryGateways {
			gatewaysMap[method] = true
		}
	}

	// Convert to slice and sort by priority
	gateways := make([]string, 0, len(gatewaysMap))
	for method := range gatewaysMap {
		gateways = append(gateways, method)
	}

	// Sort by priority
	r.sortByPriority(gateways)

	return gateways
}

// IsGatewayAvailable checks if a gateway is available for a country
func (r *GatewayRegistry) IsGatewayAvailable(country Country, method string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check global availability
	if r.globalGateways[method] {
		return true
	}

	// Check region availability
	region := GetRegion(country)
	if regionGateways, ok := r.regionGateways[region]; ok {
		if regionGateways[method] {
			return true
		}
	}

	// Check country-specific availability
	if countryGateways, ok := r.countryGateways[country]; ok {
		if countryGateways[method] {
			return true
		}
	}

	return false
}

// GetGatewayPriority returns the priority of a gateway
func (r *GatewayRegistry) GetGatewayPriority(method string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if priority, ok := r.gatewayPriority[method]; ok {
		return priority
	}
	return 999 // Default low priority
}

// sortByPriority sorts gateways by their priority (lower number = higher priority)
func (r *GatewayRegistry) sortByPriority(gateways []string) {
	// Simple bubble sort for small lists
	n := len(gateways)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			p1 := r.gatewayPriority[gateways[j]]
			p2 := r.gatewayPriority[gateways[j+1]]
			if p1 > p2 {
				gateways[j], gateways[j+1] = gateways[j+1], gateways[j]
			}
		}
	}
}

// DefaultRegistry returns a pre-configured registry with common payment gateways
func DefaultRegistry() *GatewayRegistry {
	registry := NewGatewayRegistry()

	// Nepal-specific gateways
	registry.RegisterCountryGateway(CountryNepal, "esewa", 1)
	registry.RegisterCountryGateway(CountryNepal, "khalti", 2)
	registry.RegisterCountryGateway(CountryNepal, "imepay", 3)
	registry.RegisterCountryGateway(CountryNepal, "connectips", 4)

	// India-specific gateways
	registry.RegisterCountryGateway(CountryIndia, "razorpay", 1)
	registry.RegisterCountryGateway(CountryIndia, "paytm", 2)
	registry.RegisterCountryGateway(CountryIndia, "phonepe", 3)
	registry.RegisterCountryGateway(CountryIndia, "upi", 4)

	// Southeast Asia
	registry.RegisterCountryGateway(CountrySingapore, "grab-pay", 1)
	registry.RegisterCountryGateway(CountryMalaysia, "grab-pay", 1)
	registry.RegisterCountryGateway(CountryThailand, "promptpay", 1)
	registry.RegisterCountryGateway(CountryIndonesia, "gopay", 1)
	registry.RegisterCountryGateway(CountryPhilippines, "gcash", 1)

	// Global gateways (available everywhere)
	registry.RegisterGlobalGateway("stripe", 10)
	registry.RegisterGlobalGateway("paypal", 11)
	registry.RegisterGlobalGateway("wise", 12)

	// Region-specific gateways
	registry.RegisterRegionGateway(RegionEurope, "sepa", 5)
	registry.RegisterRegionGateway(RegionNorthAmerica, "venmo", 5)
	registry.RegisterRegionGateway(RegionAfrica, "mpesa", 1)
	registry.RegisterRegionGateway(RegionLatinAmerica, "mercadopago", 1)

	return registry
}

// GatewayRecommendation provides information about recommended gateways
type GatewayRecommendation struct {
	Method      string `json:"method"`
	Priority    int    `json:"priority"`
	Scope       string `json:"scope"` // "country", "region", or "global"
	Available   bool   `json:"available"`
	Recommended bool   `json:"recommended"`
}

// GetRecommendations returns gateway recommendations for a country
func (r *GatewayRegistry) GetRecommendations(country Country) []GatewayRecommendation {
	r.mu.RLock()
	defer r.mu.RUnlock()

	recommendations := []GatewayRecommendation{}
	seenMethods := make(map[string]bool)

	// Country-specific gateways (highest priority)
	if countryGateways, ok := r.countryGateways[country]; ok {
		for method := range countryGateways {
			if !seenMethods[method] {
				recommendations = append(recommendations, GatewayRecommendation{
					Method:      method,
					Priority:    r.gatewayPriority[method],
					Scope:       "country",
					Available:   true,
					Recommended: true,
				})
				seenMethods[method] = true
			}
		}
	}

	// Region gateways
	region := GetRegion(country)
	if regionGateways, ok := r.regionGateways[region]; ok {
		for method := range regionGateways {
			if !seenMethods[method] {
				recommendations = append(recommendations, GatewayRecommendation{
					Method:      method,
					Priority:    r.gatewayPriority[method],
					Scope:       "region",
					Available:   true,
					Recommended: len(recommendations) < 5, // Recommend top 5
				})
				seenMethods[method] = true
			}
		}
	}

	// Global gateways
	for method := range r.globalGateways {
		if !seenMethods[method] {
			recommendations = append(recommendations, GatewayRecommendation{
				Method:      method,
				Priority:    r.gatewayPriority[method],
				Scope:       "global",
				Available:   true,
				Recommended: false,
			})
			seenMethods[method] = true
		}
	}

	// Sort by priority
	r.sortRecommendations(recommendations)

	return recommendations
}

func (r *GatewayRegistry) sortRecommendations(recs []GatewayRecommendation) {
	n := len(recs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if recs[j].Priority > recs[j+1].Priority {
				recs[j], recs[j+1] = recs[j+1], recs[j]
			}
		}
	}
}

// ValidateGatewayForCountry validates if a gateway can be used for a country
func (r *GatewayRegistry) ValidateGatewayForCountry(country Country, method string) error {
	if !r.IsGatewayAvailable(country, method) {
		return fmt.Errorf("gateway %s is not available for country %s", method, country)
	}
	return nil
}
