package tunnel

import (
	"fmt"
	"strings"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

// HeaderConfig contains header configuration for FRP tunnels
type HeaderConfig struct {
	// HostHeaderRewrite rewrites the Host header to a specific value
	// Empty string means no rewrite (transparent mode)
	HostHeaderRewrite string

	// RequestHeaders are custom headers to add to requests sent to the backend
	RequestHeaders map[string]string

	// ResponseHeaders are custom headers to add to responses sent to the client
	ResponseHeaders map[string]string
}

// NewHeaderConfig creates a new HeaderConfig with empty values (transparent mode)
func NewHeaderConfig() *HeaderConfig {
	return &HeaderConfig{
		RequestHeaders:  make(map[string]string),
		ResponseHeaders: make(map[string]string),
	}
}

// ParseHeaderFlags parses command-line header flags into a HeaderConfig
// Returns nil if no headers are configured (transparent mode)
func ParseHeaderFlags(
	requestHeaders []string,
	responseHeaders []string,
	hostHeader string,
) (*HeaderConfig, error) {
	// If no header flags provided, return nil for transparent mode
	if len(requestHeaders) == 0 && len(responseHeaders) == 0 && hostHeader == "" {
		return nil, nil
	}

	config := NewHeaderConfig()

	// Parse request headers
	for _, h := range requestHeaders {
		name, value, err := parseHeader(h)
		if err != nil {
			return nil, fmt.Errorf("invalid request header %q: %w", h, err)
		}
		config.RequestHeaders[name] = value
	}

	// Parse response headers
	for _, h := range responseHeaders {
		name, value, err := parseHeader(h)
		if err != nil {
			return nil, fmt.Errorf("invalid response header %q: %w", h, err)
		}
		config.ResponseHeaders[name] = value
	}

	// Set host header rewrite if provided
	if hostHeader != "" {
		config.HostHeaderRewrite = hostHeader
	}

	return config, nil
}

// parseHeader parses a header string in "name:value" format
func parseHeader(header string) (string, string, error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'name:value', got %q", header)
	}

	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if name == "" {
		return "", "", fmt.Errorf("header name cannot be empty")
	}

	return name, value, nil
}

// ApplyToProxyConfig applies this HeaderConfig to an FRP HTTPProxyConfig
func (c *HeaderConfig) ApplyToProxyConfig(proxy *v1.HTTPProxyConfig) {
	if c == nil {
		// Transparent mode - no changes
		return
	}

	// Apply host header rewrite
	if c.HostHeaderRewrite != "" {
		proxy.HostHeaderRewrite = c.HostHeaderRewrite
	}

	// Apply request headers
	if len(c.RequestHeaders) > 0 {
		proxy.RequestHeaders = v1.HeaderOperations{
			Set: c.RequestHeaders,
		}
	}

	// Apply response headers
	if len(c.ResponseHeaders) > 0 {
		proxy.ResponseHeaders = v1.HeaderOperations{
			Set: c.ResponseHeaders,
		}
	}
}

// String returns a human-readable representation of the header config
func (c *HeaderConfig) String() string {
	if c == nil {
		return "transparent mode (no header changes)"
	}

	var parts []string

	if c.HostHeaderRewrite != "" {
		parts = append(parts, fmt.Sprintf("Host: %s", c.HostHeaderRewrite))
	}

	if len(c.RequestHeaders) > 0 {
		parts = append(parts, fmt.Sprintf("%d request header(s)", len(c.RequestHeaders)))
	}

	if len(c.ResponseHeaders) > 0 {
		parts = append(parts, fmt.Sprintf("%d response header(s)", len(c.ResponseHeaders)))
	}

	if len(parts) == 0 {
		return "transparent mode (no header changes)"
	}

	return strings.Join(parts, ", ")
}
