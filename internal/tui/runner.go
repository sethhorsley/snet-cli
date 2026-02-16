package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TUIWrapper wraps the Bubble Tea program for easier integration
type TUIWrapper struct {
	program  *tea.Program
	model    Model
	quitChan chan struct{} // Signal when TUI quits
}

// NewTUIWrapper creates a new TUI wrapper
func NewTUIWrapper(
	tunnelName string,
	mainURL string,
	wildcardURL string,
	localURL string,
	accountName string,
	region string,
	version string,
	isWildcard bool,
	isReconnected bool,
	latency time.Duration,
	headerSummary string,
	hostHeaderRewrite string,
	requestHeaders map[string]string,
	responseHeaders map[string]string,
) *TUIWrapper {
	model := New(
		tunnelName,
		mainURL,
		wildcardURL,
		localURL,
		accountName,
		region,
		version,
		isWildcard,
		isReconnected,
		latency,
		headerSummary,
		hostHeaderRewrite,
		requestHeaders,
		responseHeaders,
	)

	return &TUIWrapper{
		model:    model,
		quitChan: make(chan struct{}),
	}
}

// Start starts the TUI program in a separate goroutine
func (w *TUIWrapper) Start() error {
	w.program = tea.NewProgram(w.model, tea.WithAltScreen())

	// Run in goroutine and signal when it quits
	go func() {
		if _, err := w.program.Run(); err != nil {
			// Log error but don't crash
		}
		// Signal that TUI has quit
		close(w.quitChan)
	}()

	return nil
}

// Stop stops the TUI program
func (w *TUIWrapper) Stop() {
	if w.program != nil {
		w.program.Quit()
	}
}

// QuitChan returns a channel that closes when the TUI quits
func (w *TUIWrapper) QuitChan() <-chan struct{} {
	return w.quitChan
}

// UpdateStatus updates the connection status
func (w *TUIWrapper) UpdateStatus(status string) {
	if w.program != nil {
		w.program.Send(StatusUpdateMsg{Status: status})
	}
}

// UpdateConnectionStats updates connection statistics
func (w *TUIWrapper) UpdateConnectionStats(total, open int, rt1, rt5, p50, p90 time.Duration) {
	if w.program != nil {
		w.program.Send(ConnectionStatsMsg{
			Total: total,
			Open:  open,
			RT1:   rt1,
			RT5:   rt5,
			P50:   p50,
			P90:   p90,
		})
	}
}

// LogHTTPRequest logs an HTTP request
func (w *TUIWrapper) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, isAppLog bool) {
	if w.program != nil {
		w.program.Send(HTTPRequestMsg{
			Request: HTTPRequest{
				Timestamp:  time.Now(),
				Method:     method,
				Path:       path,
				StatusCode: statusCode,
				Duration:   duration,
				IsAppLog:   isAppLog,
			},
		})
	}
}

// UpdateTunnelInfo updates tunnel information after async load
func (w *TUIWrapper) UpdateTunnelInfo(url, wildcardURL, accountName string, isWildcard, reconnected bool) {
	if w.program != nil {
		w.program.Send(TunnelLoadedMsg{
			URL:         url,
			WildcardURL: wildcardURL,
			AccountName: accountName,
			IsWildcard:  isWildcard,
			Reconnected: reconnected,
		})
	}
}
