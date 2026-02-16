package tunnel

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPProxy is a simple HTTP proxy that logs requests
type HTTPProxy struct {
	targetHost   string
	targetPort   int
	listenPort   int
	server       *http.Server
	tuiCallbacks TUICallbacks
}

// NewHTTPProxy creates a new HTTP proxy
func NewHTTPProxy(targetHost string, targetPort int, listenPort int, tuiCallbacks TUICallbacks) *HTTPProxy {
	return &HTTPProxy{
		targetHost:   targetHost,
		targetPort:   targetPort,
		listenPort:   listenPort,
		tuiCallbacks: tuiCallbacks,
	}
}

// Start starts the HTTP proxy server
func (p *HTTPProxy) Start() error {
	// Create proxy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Build target URL
		targetURL := &url.URL{
			Scheme:   "http",
			Host:     fmt.Sprintf("%s:%d", p.targetHost, p.targetPort),
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}

		// Create new request to target
		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Forward request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Failed to proxy request", http.StatusBadGateway)
			duration := time.Since(startTime)
			if p.tuiCallbacks != nil {
				p.tuiCallbacks.LogHTTPRequest(r.Method, r.URL.Path, http.StatusBadGateway, duration, true)
			}
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Write status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)

		// Log request
		duration := time.Since(startTime)
		if p.tuiCallbacks != nil {
			p.tuiCallbacks.LogHTTPRequest(r.Method, r.URL.Path, resp.StatusCode, duration, true)
		}
	})

	// Create HTTP server
	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.listenPort),
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error
		}
	}()

	return nil
}

// Stop stops the HTTP proxy server
func (p *HTTPProxy) Stop() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}
