package region

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// DetectClosestRegion detects the closest Fly.io region using Fly.io's edge routing
//
// This uses debug.fly.dev which leverages Fly.io's anycast network to automatically
// route the request to the optimal region based on network topology and latency.
// This is more accurate than manual latency testing because it uses Fly.io's actual
// routing logic which considers factors beyond simple geographic distance.
//
// The response includes a "Fly-Region" header indicating which region handled the request.
// This is the region that will provide the best performance for the user.
func DetectClosestRegion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://debug.fly.dev", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query Fly.io edge: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body to get the Fly-Region header info
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response to find the Fly-Region line
	// Response format includes "Fly-Region: <region-code>"
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Fly-Region: ") {
			region := strings.TrimSpace(strings.TrimPrefix(line, "Fly-Region: "))
			if region != "" {
				return region, nil
			}
		}
	}

	// Also check the header directly as fallback
	region := resp.Header.Get("Fly-Region")
	if region != "" {
		return region, nil
	}

	// If we can't detect the region, default to ord (Chicago - central US)
	return "ord", fmt.Errorf("could not detect region from Fly.io edge")
}

// GetRegionName returns the human-readable name for a region code
func GetRegionName(code string) string {
	names := map[string]string{
		"iad": "Ashburn, Virginia (US)",
		"ord": "Chicago, Illinois (US)",
		"sjc": "San Jose, California (US)",
		"dfw": "Dallas, Texas (US)",
		"lax": "Los Angeles, California (US)",
		"ewr": "Secaucus, New Jersey (US)",
		"yyz": "Toronto, Canada",
		"lhr": "London, UK",
		"fra": "Frankfurt, Germany",
		"ams": "Amsterdam, Netherlands",
		"cdg": "Paris, France",
		"arn": "Stockholm, Sweden",
		"nrt": "Tokyo, Japan",
		"sin": "Singapore",
		"syd": "Sydney, Australia",
		"bom": "Mumbai, India",
		"gru": "São Paulo, Brazil",
		"jnb": "Johannesburg, South Africa",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// RegionWithLatency represents a region with its measured latency
type RegionWithLatency struct {
	Code      string
	Name      string
	LatencyMS int64
}

// DetectClosestRegions detects the closest regions by measuring latency to multiple Fly.io regions
// Returns a sorted list of regions by latency (fastest first)
func DetectClosestRegions(count int) ([]RegionWithLatency, error) {
	// Test these major Fly.io regions
	testRegions := []string{
		"iad", // Ashburn, Virginia (US East)
		"ord", // Chicago, Illinois (US Central)
		"sjc", // San Jose, California (US West)
		"lhr", // London, UK
		"fra", // Frankfurt, Germany
		"nrt", // Tokyo, Japan
		"syd", // Sydney, Australia
	}

	results := make(chan RegionWithLatency, len(testRegions))

	// Measure latency to each region concurrently
	for _, regionCode := range testRegions {
		go func(code string) {
			latency, err := MeasureRegionLatency(code)
			if err != nil {
				// On error, use max latency so it won't be selected
				latency = 10000
			}
			results <- RegionWithLatency{
				Code:      code,
				Name:      GetRegionName(code),
				LatencyMS: latency,
			}
		}(regionCode)
	}

	// Collect all results
	var regions []RegionWithLatency
	for i := 0; i < len(testRegions); i++ {
		regions = append(regions, <-results)
	}

	// Sort by latency (fastest first)
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].LatencyMS < regions[j].LatencyMS
	})

	// Return top N regions
	if count > len(regions) {
		count = len(regions)
	}
	return regions[:count], nil
}

// MeasureRegionLatency measures the round-trip latency to a Fly.io region
// This uses debug.fly.dev with the Fly-Prefer-Region header to route to specific regions
func MeasureRegionLatency(regionCode string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use debug.fly.dev with region preference header
	url := "http://debug.fly.dev"

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	// Request routing to this specific region
	req.Header.Set("Fly-Prefer-Region", regionCode)

	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	_, err = client.Do(req)
	if err != nil {
		return 0, err
	}

	latencyMS := time.Since(start).Milliseconds()
	return latencyMS, nil
}
