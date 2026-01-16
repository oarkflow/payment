package payment

// Region represents a geographical region
type Region string

const (
	RegionSouthAsia     Region = "south-asia"
	RegionSoutheastAsia Region = "southeast-asia"
	RegionEastAsia      Region = "east-asia"
	RegionNorthAmerica  Region = "north-america"
	RegionEurope        Region = "europe"
	RegionMiddleEast    Region = "middle-east"
	RegionAfrica        Region = "africa"
	RegionOceania       Region = "oceania"
	RegionLatinAmerica  Region = "latin-america"
	RegionGlobal        Region = "global"
)

// Country represents a country code (ISO 3166-1 alpha-2)
type Country string

const (
	// South Asia
	CountryNepal      Country = "NP"
	CountryIndia      Country = "IN"
	CountryPakistan   Country = "PK"
	CountryBangladesh Country = "BD"
	CountrySriLanka   Country = "LK"

	// Southeast Asia
	CountrySingapore  Country = "SG"
	CountryMalaysia   Country = "MY"
	CountryIndonesia  Country = "ID"
	CountryThailand   Country = "TH"
	CountryPhilippines Country = "PH"
	CountryVietnam    Country = "VN"

	// East Asia
	CountryChina      Country = "CN"
	CountryJapan      Country = "JP"
	CountrySouthKorea Country = "KR"

	// North America
	CountryUSA    Country = "US"
	CountryCanada Country = "CA"
	CountryMexico Country = "MX"

	// Europe
	CountryUK      Country = "GB"
	CountryGermany Country = "DE"
	CountryFrance  Country = "FR"
	CountrySpain   Country = "ES"
	CountryItaly   Country = "IT"

	// Middle East
	CountryUAE         Country = "AE"
	CountrySaudiArabia Country = "SA"

	// Africa
	CountryNigeria    Country = "NG"
	CountrySouthAfrica Country = "ZA"
	CountryKenya      Country = "KE"

	// Oceania
	CountryAustralia  Country = "AU"
	CountryNewZealand Country = "NZ"

	// Latin America
	CountryBrazil    Country = "BR"
	CountryArgentina Country = "AR"

	// Global
	CountryGlobal Country = "GLOBAL"
)

// RegionInfo contains metadata about a region
type RegionInfo struct {
	Region       Region    `json:"region"`
	Countries    []Country `json:"countries"`
	DefaultCurrency string `json:"default_currency"`
	Description  string    `json:"description"`
}

// CountryInfo contains metadata about a country
type CountryInfo struct {
	Country         Country  `json:"country"`
	Region          Region   `json:"region"`
	Name            string   `json:"name"`
	Currency        string   `json:"currency"`
	SupportedMethods []string `json:"supported_methods"`
}

// RegionMap maps regions to their countries
var RegionMap = map[Region][]Country{
	RegionSouthAsia: {
		CountryNepal, CountryIndia, CountryPakistan,
		CountryBangladesh, CountrySriLanka,
	},
	RegionSoutheastAsia: {
		CountrySingapore, CountryMalaysia, CountryIndonesia,
		CountryThailand, CountryPhilippines, CountryVietnam,
	},
	RegionEastAsia: {
		CountryChina, CountryJapan, CountrySouthKorea,
	},
	RegionNorthAmerica: {
		CountryUSA, CountryCanada, CountryMexico,
	},
	RegionEurope: {
		CountryUK, CountryGermany, CountryFrance,
		CountrySpain, CountryItaly,
	},
	RegionMiddleEast: {
		CountryUAE, CountrySaudiArabia,
	},
	RegionAfrica: {
		CountryNigeria, CountrySouthAfrica, CountryKenya,
	},
	RegionOceania: {
		CountryAustralia, CountryNewZealand,
	},
	RegionLatinAmerica: {
		CountryBrazil, CountryArgentina,
	},
}

// CountryToRegion maps countries to their regions
var CountryToRegion = map[Country]Region{
	// South Asia
	CountryNepal:      RegionSouthAsia,
	CountryIndia:      RegionSouthAsia,
	CountryPakistan:   RegionSouthAsia,
	CountryBangladesh: RegionSouthAsia,
	CountrySriLanka:   RegionSouthAsia,

	// Southeast Asia
	CountrySingapore:   RegionSoutheastAsia,
	CountryMalaysia:    RegionSoutheastAsia,
	CountryIndonesia:   RegionSoutheastAsia,
	CountryThailand:    RegionSoutheastAsia,
	CountryPhilippines: RegionSoutheastAsia,
	CountryVietnam:     RegionSoutheastAsia,

	// East Asia
	CountryChina:      RegionEastAsia,
	CountryJapan:      RegionEastAsia,
	CountrySouthKorea: RegionEastAsia,

	// North America
	CountryUSA:    RegionNorthAmerica,
	CountryCanada: RegionNorthAmerica,
	CountryMexico: RegionNorthAmerica,

	// Europe
	CountryUK:      RegionEurope,
	CountryGermany: RegionEurope,
	CountryFrance:  RegionEurope,
	CountrySpain:   RegionEurope,
	CountryItaly:   RegionEurope,

	// Middle East
	CountryUAE:         RegionMiddleEast,
	CountrySaudiArabia: RegionMiddleEast,

	// Africa
	CountryNigeria:     RegionAfrica,
	CountrySouthAfrica: RegionAfrica,
	CountryKenya:       RegionAfrica,

	// Oceania
	CountryAustralia:  RegionOceania,
	CountryNewZealand: RegionOceania,

	// Latin America
	CountryBrazil:    RegionLatinAmerica,
	CountryArgentina: RegionLatinAmerica,
}

// GetRegion returns the region for a given country
func GetRegion(country Country) Region {
	if region, ok := CountryToRegion[country]; ok {
		return region
	}
	return RegionGlobal
}

// GetCountriesInRegion returns all countries in a given region
func GetCountriesInRegion(region Region) []Country {
	if countries, ok := RegionMap[region]; ok {
		return countries
	}
	return []Country{}
}
