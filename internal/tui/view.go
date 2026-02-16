package tui

import (
	"fmt"
	"strings"
	"time"
)

// renderMain renders the main TUI view
func renderMain(m Model) string {
	// Build content sections
	var content strings.Builder

	// Header
	content.WriteString(renderHeader(m))
	content.WriteString("\n")

	// Session info
	content.WriteString(renderSessionInfo(m))
	content.WriteString("\n")

	// Forwarding URLs
	content.WriteString(renderForwarding(m))
	content.WriteString("\n")

	// Connection stats
	content.WriteString(renderStats(m))
	content.WriteString("\n")

	// HTTP request log (fills remaining space)
	content.WriteString(renderRequestLog(m))

	// Build final output with footer at bottom
	var output strings.Builder

	// Content
	contentStr := content.String()
	output.WriteString(contentStr)

	// Calculate padding to push footer to bottom
	contentLines := strings.Count(contentStr, "\n")
	footerLines := 2 // Separator + shortcuts
	availableHeight := m.Height
	if availableHeight < 10 {
		availableHeight = 24 // Default height
	}

	paddingLines := availableHeight - contentLines - footerLines - 1
	if paddingLines > 0 {
		output.WriteString(strings.Repeat("\n", paddingLines))
	}

	// Footer pinned to bottom
	output.WriteString("\n")
	output.WriteString(renderFooter(m))

	return output.String()
}

// renderHeader renders the header section
func renderHeader(m Model) string {
	var b strings.Builder

	// Version at the very top
	b.WriteString(LabelStyle.Render("Version:"))
	b.WriteString(ValueStyle.Render(m.Version))
	b.WriteString("\n\n")

	// Title and status
	title := "snet - Self-Hosted Tunnels"
	statusText := strings.ToUpper(m.Status)
	statusStyle := GetStatusTextStyle(m.Status)
	status := statusStyle.Render(statusText)

	b.WriteString(fmt.Sprintf("%s    %s",
		HeaderStyle.Render(title),
		status,
	))

	return b.String()
}

// renderSessionInfo renders session information
func renderSessionInfo(m Model) string {
	var b strings.Builder

	// Account
	b.WriteString(LabelStyle.Render("Account:"))
	b.WriteString(ValueStyle.Render(fmt.Sprintf("%s (%s)", m.AccountName, m.AccountPlan)))
	b.WriteString("\n")

	// Region
	b.WriteString(LabelStyle.Render("Region:"))
	regionText := fmt.Sprintf("%s (%s)", m.Region, formatDuration(m.Latency))
	b.WriteString(ValueStyle.Render(regionText))
	b.WriteString("\n")

	// Tunnel name
	tunnelName := m.TunnelName
	if m.IsReconnected {
		tunnelName += " (reconnected)"
	}
	b.WriteString(LabelStyle.Render("Tunnel:"))
	b.WriteString(ValueStyle.Render(tunnelName))

	// Header configuration (if any) - compact view
	hasHeaders := m.HostHeaderRewrite != "" || len(m.RequestHeaders) > 0 || len(m.ResponseHeaders) > 0
	if hasHeaders {
		b.WriteString("\n")

		// Count total headers
		headerCount := 0
		if m.HostHeaderRewrite != "" {
			headerCount++
		}
		headerCount += len(m.RequestHeaders)
		headerCount += len(m.ResponseHeaders)

		// Show compact header summary
		if m.ShowHeaderDetails {
			b.WriteString(LabelStyle.Render("Headers:"))
			b.WriteString(ValueStyle.Render(fmt.Sprintf("%d set", headerCount)))
			b.WriteString(ShortcutDescStyle.Render(" (press 'h' to hide)"))
			b.WriteString("\n")

			// Host header rewrite
			if m.HostHeaderRewrite != "" {
				b.WriteString(LabelStyle.Render("  Host →"))
				b.WriteString(ValueStyle.Render(m.HostHeaderRewrite))
				b.WriteString("\n")
			}

			// Request headers
			if len(m.RequestHeaders) > 0 {
				b.WriteString(LabelStyle.Render("  Request:"))
				b.WriteString("\n")
				for name, value := range m.RequestHeaders {
					b.WriteString(LabelStyle.Render(fmt.Sprintf("    %s:", name)))
					b.WriteString(ValueStyle.Render(value))
					b.WriteString("\n")
				}
			}

			// Response headers
			if len(m.ResponseHeaders) > 0 {
				b.WriteString(LabelStyle.Render("  Response:"))
				b.WriteString("\n")
				for name, value := range m.ResponseHeaders {
					b.WriteString(LabelStyle.Render(fmt.Sprintf("    %s:", name)))
					b.WriteString(ValueStyle.Render(value))
					b.WriteString("\n")
				}
			}
		} else {
			b.WriteString(LabelStyle.Render("Headers:"))
			b.WriteString(ValueStyle.Render(fmt.Sprintf("%d set", headerCount)))
			b.WriteString(ShortcutDescStyle.Render(" (press 'h' to show)"))
		}
	}

	return b.String()
}

// renderForwarding renders forwarding URLs
func renderForwarding(m Model) string {
	var b strings.Builder

	// Main forwarding with arrow format (aligned with other values)
	b.WriteString(LabelStyle.Render("Forwarding:"))
	b.WriteString(URLStyle.Render(m.MainURL))
	b.WriteString(ValueStyle.Render(" -> "))
	b.WriteString(ValueStyle.Render(m.LocalURL))
	b.WriteString("\n")

	// Wildcard forwarding (if enabled)
	if m.IsWildcard && m.WildcardURL != "" {
		b.WriteString(LabelStyle.Render("Wildcard:"))
		b.WriteString(URLStyle.Render(m.WildcardURL))
		b.WriteString(ValueStyle.Render(" -> "))
		b.WriteString(ValueStyle.Render(m.LocalURL))
		b.WriteString("\n")
	}

	// Web interface
	b.WriteString(LabelStyle.Render("Web Interface:"))
	b.WriteString(URLStyle.Render("http://127.0.0.1:4040"))

	return b.String()
}

// renderStats renders connection statistics
func renderStats(m Model) string {
	var b strings.Builder

	// Connections label
	b.WriteString(LabelStyle.Render("Connections:"))
	b.WriteString(ValueStyle.Render("ttl     opn"))
	b.WriteString("\n")

	// Values under the labels, aligned right (20 spaces for label column)
	b.WriteString(LabelStyle.Render(""))
	b.WriteString(ValueStyle.Render(fmt.Sprintf("%-7d %d", m.TotalConnections, m.OpenConnections)))

	// Response times (only if we have connections)
	if m.TotalConnections > 0 {
		b.WriteString("\n")
		b.WriteString(LabelStyle.Render("Response Times:"))
		b.WriteString(ValueStyle.Render(fmt.Sprintf(
			"rt1=%s  rt5=%s  p50=%s  p90=%s",
			formatDuration(m.RT1),
			formatDuration(m.RT5),
			formatDuration(m.P50),
			formatDuration(m.P90),
		)))
	}

	return b.String()
}

// renderRequestLog renders the HTTP request log
func renderRequestLog(m Model) string {
	var b strings.Builder

	// Header with filter indicator
	header := "HTTP Requests"
	if !m.ShowAllLogs {
		header += " (app logs only - press 't' to show all)"
	} else {
		header += " (all logs - press 't' to filter)"
	}
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Table header
	b.WriteString(renderRequestTableHeader())
	b.WriteString("\n")

	// Filter and render requests
	requests := m.Requests
	if !m.ShowAllLogs {
		requests = filterAppLogs(requests)
	}

	// Show last N requests that fit in the terminal
	maxRows := m.Height - 20 // Reserve space for header and stats
	if maxRows < 5 {
		maxRows = 5
	}

	startIdx := 0
	if len(requests) > maxRows {
		startIdx = len(requests) - maxRows
	}

	for i := startIdx; i < len(requests); i++ {
		b.WriteString(renderRequestRow(requests[i], i%2 == 0))
		b.WriteString("\n")
	}

	if len(requests) == 0 {
		b.WriteString(TableRowAltStyle.Render("  No requests yet..."))
		b.WriteString("\n")
	}

	return b.String()
}

// renderRequestTableHeader renders the table header
func renderRequestTableHeader() string {
	return fmt.Sprintf("%-8s  %-6s  %-30s  %-30s  %s",
		"TIME",
		"METHOD",
		"PATH",
		"STATUS",
		"DURATION",
	)
}

// renderRequestRow renders a single request row
func renderRequestRow(req HTTPRequest, alt bool) string {
	// Base style
	rowStyle := TableRowStyle
	if alt {
		rowStyle = TableRowAltStyle
	}

	// Format time
	timeStr := req.Timestamp.Format("15:04:05")

	// Method with color
	methodStyle := GetMethodStyle(req.Method)
	methodStr := methodStyle.Render(fmt.Sprintf("%-6s", req.Method))

	// Path (truncate if needed)
	path := req.Path
	if len(path) > 30 {
		path = path[:27] + "..."
	}
	pathStr := rowStyle.Render(fmt.Sprintf("%-30s", path))

	// Status code with color and text (fixed width)
	statusStyle := GetStatusStyle(req.StatusCode)
	statusText := getStatusText(req.StatusCode)
	statusFull := fmt.Sprintf("%d %s", req.StatusCode, statusText)
	// Pad or truncate to 30 characters
	if len(statusFull) > 30 {
		statusFull = statusFull[:27] + "..."
	}
	statusStr := statusStyle.Render(fmt.Sprintf("%-30s", statusFull))

	// Duration
	durationStr := rowStyle.Render(formatDuration(req.Duration))

	return fmt.Sprintf("%-8s  %s  %s  %s  %s",
		rowStyle.Render(timeStr),
		methodStr,
		pathStr,
		statusStr,
		durationStr,
	)
}

// renderFooter renders the footer with keyboard shortcuts
func renderFooter(m Model) string {
	var b strings.Builder

	// Add separator line
	width := m.Width
	if width < 40 {
		width = 80 // Default width
	}
	separator := strings.Repeat("─", width)
	b.WriteString(DividerStyle.Render(separator))
	b.WriteString("\n")

	// Show quit confirmation or normal shortcuts
	if m.ConfirmQuit {
		// Show quit confirmation message
		quitMsg := fmt.Sprintf("%s  %s  %s",
			StatusConnectingStyle.Render("⚠ Press 'q' or Ctrl+C again to quit"),
			FooterStyle.Render("•"),
			ShortcutDescStyle.Render("Press Esc to cancel"),
		)
		b.WriteString(quitMsg)
	} else {
		// Normal shortcuts
		shortcuts := []string{
			fmt.Sprintf("%s %s", ShortcutKeyStyle.Render("[t]"), ShortcutDescStyle.Render("toggle logs")),
			fmt.Sprintf("%s %s", ShortcutKeyStyle.Render("[h]"), ShortcutDescStyle.Render("toggle headers")),
			fmt.Sprintf("%s %s", ShortcutKeyStyle.Render("[c]"), ShortcutDescStyle.Render("clear")),
			fmt.Sprintf("%s %s", ShortcutKeyStyle.Render("[?]"), ShortcutDescStyle.Render("help")),
			fmt.Sprintf("%s %s", ShortcutKeyStyle.Render("[Ctrl+C]"), ShortcutDescStyle.Render("quit")),
		}
		b.WriteString(FooterStyle.Render(strings.Join(shortcuts, "  •  ")))
	}

	return b.String()
}

// renderHelp renders the help screen
func renderHelp(m Model) string {
	var b strings.Builder

	b.WriteString(HelpTitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"t", "Toggle between app logs and all tunnel logs"},
		{"h", "Toggle header details"},
		{"c", "Clear request history"},
		{"?", "Show/hide this help screen"},
		{"q, Ctrl+C", "Quit and stop tunnel"},
	}

	for _, sc := range shortcuts {
		b.WriteString(HelpKeyStyle.Render(sc.key))
		b.WriteString(HelpDescStyle.Render(sc.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(FooterStyle.Render("Press any key to return..."))

	return BoxStyle.Render(b.String())
}

// renderShutdown renders the shutdown message
func renderShutdown(m Model) string {
	msg := "Shutting down tunnel...\n\n"
	msg += "Your tunnel resources have been preserved:\n"
	msg += fmt.Sprintf("  • DNS: %s\n", m.MainURL)
	msg += "  • SSL certificate maintained\n"
	msg += "  • Database record kept\n\n"
	msg += fmt.Sprintf("To reconnect: snet <port> --name %s\n", m.TunnelName)

	return StatusConnectingStyle.Render(msg)
}

// Helper functions

func filterAppLogs(requests []HTTPRequest) []HTTPRequest {
	filtered := make([]HTTPRequest, 0)
	for _, req := range requests {
		if req.IsAppLog {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}

	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}

	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	return fmt.Sprintf("%.2fs", d.Seconds())
}

func getStatusText(statusCode int) string {
	switch statusCode {
	// 1xx Informational
	case 100:
		return "Continue"
	case 101:
		return "Switching Protocols"
	case 102:
		return "Processing"
	case 103:
		return "Early Hints"

	// 2xx Success
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 202:
		return "Accepted"
	case 203:
		return "Non-Authoritative Information"
	case 204:
		return "No Content"
	case 205:
		return "Reset Content"
	case 206:
		return "Partial Content"
	case 207:
		return "Multi-Status"
	case 208:
		return "Already Reported"
	case 226:
		return "IM Used"

	// 3xx Redirection
	case 300:
		return "Multiple Choices"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Found"
	case 303:
		return "See Other"
	case 304:
		return "Not Modified"
	case 305:
		return "Use Proxy"
	case 307:
		return "Temporary Redirect"
	case 308:
		return "Permanent Redirect"

	// 4xx Client Error
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 402:
		return "Payment Required"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 406:
		return "Not Acceptable"
	case 407:
		return "Proxy Authentication Required"
	case 408:
		return "Request Timeout"
	case 409:
		return "Conflict"
	case 410:
		return "Gone"
	case 411:
		return "Length Required"
	case 412:
		return "Precondition Failed"
	case 413:
		return "Payload Too Large"
	case 414:
		return "URI Too Long"
	case 415:
		return "Unsupported Media Type"
	case 416:
		return "Range Not Satisfiable"
	case 417:
		return "Expectation Failed"
	case 418:
		return "I'm a teapot"
	case 421:
		return "Misdirected Request"
	case 422:
		return "Unprocessable Entity"
	case 423:
		return "Locked"
	case 424:
		return "Failed Dependency"
	case 425:
		return "Too Early"
	case 426:
		return "Upgrade Required"
	case 428:
		return "Precondition Required"
	case 429:
		return "Too Many Requests"
	case 431:
		return "Request Header Fields Too Large"
	case 451:
		return "Unavailable For Legal Reasons"

	// 5xx Server Error
	case 500:
		return "Internal Server Error"
	case 501:
		return "Not Implemented"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	case 505:
		return "HTTP Version Not Supported"
	case 506:
		return "Variant Also Negotiates"
	case 507:
		return "Insufficient Storage"
	case 508:
		return "Loop Detected"
	case 510:
		return "Not Extended"
	case 511:
		return "Network Authentication Required"

	default:
		return ""
	}
}
