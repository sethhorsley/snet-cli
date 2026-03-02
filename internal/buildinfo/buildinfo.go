package buildinfo

// Build-time variables set via ldflags
var (
	// Version is the semantic version (set via -ldflags)
	Version = "dev"

	// Mode is "development" or "production" (set via -ldflags)
	Mode = "development"

	// APIBase is the API endpoint (set via -ldflags)
	// Defaults based on Mode if not explicitly set
	APIBase = ""
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"

	DevAPIBase  = "http://localhost:3001/api/v1"
	ProdAPIBase = "https://snet.dev/api/v1"
)

// GetAPIBase returns the API base URL, using the appropriate default if not set
func GetAPIBase() string {
	if APIBase != "" {
		return APIBase
	}
	if Mode == ModeProduction {
		return ProdAPIBase
	}
	return DevAPIBase
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return Mode == ModeDevelopment
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return Mode == ModeProduction
}
