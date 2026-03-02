package tunnel

import (
	"bytes"
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
	webInspector *WebInspector
}

// NewHTTPProxy creates a new HTTP proxy
func NewHTTPProxy(targetHost string, targetPort int, listenPort int, tuiCallbacks TUICallbacks, webInspector *WebInspector) *HTTPProxy {
	return &HTTPProxy{
		targetHost:   targetHost,
		targetPort:   targetPort,
		listenPort:   listenPort,
		tuiCallbacks: tuiCallbacks,
		webInspector: webInspector,
	}
}

// Start starts the HTTP proxy server
func (p *HTTPProxy) Start() error {
	// Create proxy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Read request body
		var requestBody bytes.Buffer
		if r.Body != nil {
			io.Copy(&requestBody, r.Body)
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(requestBody.Bytes()))
		}

		// Build target URL
		targetURL := &url.URL{
			Scheme:   "http",
			Host:     fmt.Sprintf("%s:%d", p.targetHost, p.targetPort),
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}

		// Create new request to target
		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), bytes.NewReader(requestBody.Bytes()))
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

		// Preserve original Host header from X-Forwarded-Host if present
		// This makes snet behave like ngrok where the Host header matches the tunnel URL
		if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
			proxyReq.Host = forwardedHost
			proxyReq.Header.Set("Host", forwardedHost)
		}

		// Forward request
		client := &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects - let the client handle them
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Do(proxyReq)
		if err != nil {
			duration := time.Since(startTime)
			errorMsg := fmt.Sprintf("Failed to proxy request: %v", err)

			// Return error response to client
			http.Error(w, errorMsg, http.StatusBadGateway)

			// Log to TUI
			if p.tuiCallbacks != nil {
				p.tuiCallbacks.LogHTTPRequest(r.Method, r.URL.Path, http.StatusBadGateway, duration, true)
			}

			// Log detailed error to web inspector
			if p.webInspector != nil {
				// Determine the host being used
				host := proxyReq.Host
				if host == "" {
					host = r.Host
				}

				p.webInspector.AddRequest(InspectorRequest{
					ID:             fmt.Sprintf("%d", time.Now().UnixNano()),
					Timestamp:      startTime,
					Method:         r.Method,
					Path:           r.URL.Path,
					Host:           host,
					FullURL:        r.URL.String(),
					StatusCode:     http.StatusBadGateway,
					StatusText:     "Bad Gateway",
					Duration:       duration,
					RequestHeaders: r.Header,
					ResponseHeaders: map[string][]string{
						"Content-Type": {"text/plain; charset=utf-8"},
					},
					RequestBody:   requestBody.Bytes(),
					ResponseBody:  []byte(errorMsg),
					RemoteAddr:    r.RemoteAddr,
					ContentType:   "text/plain; charset=utf-8",
					ContentLength: int64(len(errorMsg)),
					Error:         err.Error(),
					TargetURL:     targetURL.String(),
				})
			}
			return
		}
		defer resp.Body.Close()

		// Read response body
		var responseBody bytes.Buffer
		io.Copy(&responseBody, resp.Body)

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Write status code
		w.WriteHeader(resp.StatusCode)

		// Write response body
		w.Write(responseBody.Bytes())

		// Log request to TUI
		duration := time.Since(startTime)
		if p.tuiCallbacks != nil {
			p.tuiCallbacks.LogHTTPRequest(r.Method, r.URL.Path, resp.StatusCode, duration, true)
		}

		// Log to web inspector
		if p.webInspector != nil {
			statusText := http.StatusText(resp.StatusCode)

			// Determine the host being used
			host := proxyReq.Host
			if host == "" {
				host = r.Host
			}

			p.webInspector.AddRequest(InspectorRequest{
				ID:              fmt.Sprintf("%d", time.Now().UnixNano()),
				Timestamp:       startTime,
				Method:          r.Method,
				Path:            r.URL.Path,
				Host:            host,
				FullURL:         r.URL.String(),
				StatusCode:      resp.StatusCode,
				StatusText:      statusText,
				Duration:        duration,
				RequestHeaders:  r.Header,
				ResponseHeaders: resp.Header,
				RequestBody:     requestBody.Bytes(),
				ResponseBody:    responseBody.Bytes(),
				RemoteAddr:      r.RemoteAddr,
				ContentType:     resp.Header.Get("Content-Type"),
				ContentLength:   resp.ContentLength,
				TargetURL:       targetURL.String(),
			})
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
