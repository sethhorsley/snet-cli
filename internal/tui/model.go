package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the application state for the TUI
type Model struct {
	// Session info
	Status      string // "online", "offline", "connecting"
	AccountName string
	AccountPlan string
	Version     string
	Region      string
	Latency     time.Duration

	// Tunnel info
	TunnelName        string
	MainURL           string
	WildcardURL       string
	IsWildcard        bool
	IsReconnected     bool
	HeaderSummary     string            // Summary of header configuration
	HostHeaderRewrite string            // Host header rewrite value
	RequestHeaders    map[string]string // Custom request headers
	ResponseHeaders   map[string]string // Custom response headers

	// Connection stats
	TotalConnections int
	OpenConnections  int
	RT1              time.Duration // Last 1 request
	RT5              time.Duration // Average of last 5
	P50              time.Duration // 50th percentile
	P90              time.Duration // 90th percentile

	// HTTP request log
	Requests    []HTTPRequest
	MaxRequests int  // Circular buffer size
	ShowAllLogs bool // Toggle between app logs and all tunnel logs

	// UI state
	Width              int
	Height             int
	ShowHelp           bool
	Quitting           bool
	ConfirmQuit        bool      // Waiting for second quit confirmation
	QuitConfirmExpires time.Time // When quit confirmation expires
}

// HTTPRequest represents a logged HTTP request
type HTTPRequest struct {
	Timestamp  time.Time
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	IsAppLog   bool // True for app logs, false for tunnel infrastructure
}

// Messages for Bubble Tea

type StatusUpdateMsg struct {
	Status string
}

type ConnectionStatsMsg struct {
	Total int
	Open  int
	RT1   time.Duration
	RT5   time.Duration
	P50   time.Duration
	P90   time.Duration
}

type HTTPRequestMsg struct {
	Request HTTPRequest
}

type TickMsg time.Time

type ClearQuitConfirmMsg struct{}

type TunnelLoadedMsg struct {
	URL         string
	WildcardURL string
	AccountName string
	IsWildcard  bool
	Reconnected bool
}

// New creates a new TUI model
func New(tunnelName, mainURL, wildcardURL, accountName, region, version string, isWildcard, isReconnected bool, latency time.Duration, headerSummary string, hostHeaderRewrite string, requestHeaders, responseHeaders map[string]string) Model {
	return Model{
		Status:            "connecting",
		AccountName:       accountName,
		AccountPlan:       "Free", // TODO: Get from API
		Version:           version,
		Region:            region,
		Latency:           latency,
		TunnelName:        tunnelName,
		MainURL:           mainURL,
		WildcardURL:       wildcardURL,
		IsWildcard:        isWildcard,
		IsReconnected:     isReconnected,
		HeaderSummary:     headerSummary,
		HostHeaderRewrite: hostHeaderRewrite,
		RequestHeaders:    requestHeaders,
		ResponseHeaders:   responseHeaders,
		Requests:          make([]HTTPRequest, 0),
		MaxRequests:       50, // Keep last 50 requests
		ShowAllLogs:       false,
		Width:             80,
		Height:            24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case StatusUpdateMsg:
		m.Status = msg.Status
		return m, nil

	case ConnectionStatsMsg:
		m.TotalConnections = msg.Total
		m.OpenConnections = msg.Open
		m.RT1 = msg.RT1
		m.RT5 = msg.RT5
		m.P50 = msg.P50
		m.P90 = msg.P90
		return m, nil

	case HTTPRequestMsg:
		m.addRequest(msg.Request)
		return m, nil

	case TickMsg:
		// Check if quit confirmation has expired
		if m.ConfirmQuit && time.Now().After(m.QuitConfirmExpires) {
			m.ConfirmQuit = false
		}
		return m, tickCmd()

	case ClearQuitConfirmMsg:
		m.ConfirmQuit = false
		return m, nil

	case TunnelLoadedMsg:
		// Update tunnel info from async load
		m.MainURL = msg.URL
		m.WildcardURL = msg.WildcardURL
		m.AccountName = msg.AccountName
		m.IsWildcard = msg.IsWildcard
		m.IsReconnected = msg.Reconnected
		m.Status = "connecting" // Will be updated to online by FRP runner
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're in quit confirmation mode
	if m.ConfirmQuit {
		switch msg.String() {
		case "ctrl+c", "q":
			// Second press - actually quit
			m.Quitting = true
			return m, tea.Quit
		case "esc":
			// Cancel quit confirmation
			m.ConfirmQuit = false
			return m, nil
		default:
			// Any other key cancels quit confirmation
			m.ConfirmQuit = false
			// Fall through to handle the key normally
		}
	}

	// Normal key handling
	switch msg.String() {
	case "ctrl+c", "q":
		// First press - show confirmation
		m.ConfirmQuit = true
		m.QuitConfirmExpires = time.Now().Add(3 * time.Second)
		return m, clearQuitConfirmAfter(3 * time.Second)

	case "t":
		m.ShowAllLogs = !m.ShowAllLogs
		return m, nil

	case "c":
		m.Requests = make([]HTTPRequest, 0)
		return m, nil

	case "?":
		m.ShowHelp = !m.ShowHelp
		return m, nil
	}

	return m, nil
}

// addRequest adds a new HTTP request to the log (circular buffer)
func (m *Model) addRequest(req HTTPRequest) {
	m.Requests = append(m.Requests, req)
	if len(m.Requests) > m.MaxRequests {
		m.Requests = m.Requests[1:]
	}
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// clearQuitConfirmAfter returns a command that clears quit confirmation after a duration
func clearQuitConfirmAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ClearQuitConfirmMsg{}
	})
}

// View renders the TUI
func (m Model) View() string {
	if m.Quitting {
		return renderShutdown(m)
	}

	if m.ShowHelp {
		return renderHelp(m)
	}

	return renderMain(m)
}
