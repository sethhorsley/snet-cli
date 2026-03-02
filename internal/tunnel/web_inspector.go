package tunnel

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

// InspectorRequest represents a captured HTTP request with full details
type InspectorRequest struct {
	ID              string              `json:"id"`
	Timestamp       time.Time           `json:"timestamp"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Host            string              `json:"host"`
	FullURL         string              `json:"full_url"`
	StatusCode      int                 `json:"status_code"`
	StatusText      string              `json:"status_text"`
	Duration        time.Duration       `json:"duration"`
	RequestHeaders  map[string][]string `json:"request_headers"`
	ResponseHeaders map[string][]string `json:"response_headers"`
	RequestBody     []byte              `json:"request_body"`
	ResponseBody    []byte              `json:"response_body"`
	RemoteAddr      string              `json:"remote_addr"`
	ContentType     string              `json:"content_type"`
	ContentLength   int64               `json:"content_length"`
	Error           string              `json:"error,omitempty"`
	TargetURL       string              `json:"target_url,omitempty"`
}

// WebInspector provides a web interface for inspecting HTTP traffic
type WebInspector struct {
	port          int
	server        *http.Server
	requests      []InspectorRequest
	requestsMutex sync.RWMutex
	maxRequests   int
	tunnelInfo    TunnelInfo
}

// TunnelInfo contains information about the tunnel
type TunnelInfo struct {
	Status   string
	MainURL  string
	LocalURL string
	Region   string
	Version  string
}

// NewWebInspector creates a new web inspector
func NewWebInspector(port int, maxRequests int, tunnelInfo TunnelInfo) *WebInspector {
	return &WebInspector{
		port:        port,
		requests:    make([]InspectorRequest, 0, maxRequests),
		maxRequests: maxRequests,
		tunnelInfo:  tunnelInfo,
	}
}

// Start starts the web inspector server
func (w *WebInspector) Start() error {
	mux := http.NewServeMux()

	// Main page
	mux.HandleFunc("/", w.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/requests", w.handleAPIRequests)
	mux.HandleFunc("/api/requests/", w.handleAPIRequestDetail)
	mux.HandleFunc("/api/tunnel-info", w.handleAPITunnelInfo)

	w.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", w.port),
		Handler: mux,
	}

	go func() {
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error
		}
	}()

	return nil
}

// Stop stops the web inspector server
func (w *WebInspector) Stop() error {
	if w.server != nil {
		return w.server.Close()
	}
	return nil
}

// AddRequest adds a request to the inspector
func (w *WebInspector) AddRequest(req InspectorRequest) {
	w.requestsMutex.Lock()
	defer w.requestsMutex.Unlock()

	// Generate ID if not set
	if req.ID == "" {
		req.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Add to beginning of slice
	w.requests = append([]InspectorRequest{req}, w.requests...)

	// Trim to max size
	if len(w.requests) > w.maxRequests {
		w.requests = w.requests[:w.maxRequests]
	}
}

// handleIndex serves the main inspection page
func (w *WebInspector) handleIndex(rw http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		http.Error(rw, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"TunnelInfo": w.tunnelInfo,
	}

	rw.Header().Set("Content-Type", "text/html")
	tmpl.Execute(rw, data)
}

// handleAPIRequests returns the list of requests as JSON
func (w *WebInspector) handleAPIRequests(rw http.ResponseWriter, r *http.Request) {
	w.requestsMutex.RLock()
	defer w.requestsMutex.RUnlock()

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(w.requests)
}

// handleAPIRequestDetail returns details for a specific request
func (w *WebInspector) handleAPIRequestDetail(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/requests/"):]

	w.requestsMutex.RLock()
	defer w.requestsMutex.RUnlock()

	for _, req := range w.requests {
		if req.ID == id {
			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(req)
			return
		}
	}

	http.NotFound(rw, r)
}

// handleAPITunnelInfo returns tunnel information
func (w *WebInspector) handleAPITunnelInfo(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(w.tunnelInfo)
}

// UpdateTunnelInfo updates the tunnel information
func (w *WebInspector) UpdateTunnelInfo(info TunnelInfo) {
	w.tunnelInfo = info
}
