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
	content.WriteString("\n\n")

	// Session info
	content.WriteString(renderSessionInfo(m))
	content.WriteString("\n\n")

	// Forwarding URLs
	content.WriteString(renderForwarding(m))
	content.WriteString("\n\n")

	// Connection stats
	content.WriteString(renderStats(m))
	content.WriteString("\n\n")

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
	title := "snet - Self-Hosted Tunnels"

	// Status indicator
	statusText := strings.ToUpper(m.Status)
	statusStyle := GetStatusTextStyle(m.Status)
	status := statusStyle.Render(statusText)

	// Combine
	header := fmt.Sprintf("%s    %s",
		HeaderStyle.Render(title),
		status,
	)

	return header
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

	// Version
	b.WriteString(LabelStyle.Render("Version:"))
	b.WriteString(ValueStyle.Render(m.Version))
	b.WriteString("\n")

	// Tunnel name
	tunnelName := m.TunnelName
	if m.IsReconnected {
		tunnelName += " (reconnected)"
	}
	b.WriteString(LabelStyle.Render("Tunnel:"))
	b.WriteString(ValueStyle.Render(tunnelName))

	return b.String()
}

// renderForwarding renders forwarding URLs
func renderForwarding(m Model) string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("Forwarding"))
	b.WriteString("\n")

	// Main URL
	b.WriteString(LabelStyle.Render("Web Interface:"))
	b.WriteString(URLStyle.Render(m.MainURL))
	b.WriteString("\n")

	// Wildcard URL (if enabled)
	if m.IsWildcard && m.WildcardURL != "" {
		b.WriteString(LabelStyle.Render("Wildcard:"))
		b.WriteString(URLStyle.Render(m.WildcardURL))
		b.WriteString("\n")
	}

	return b.String()
}

// renderStats renders connection statistics
func renderStats(m Model) string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("Connections"))
	b.WriteString("\n")

	// Total and open connections
	b.WriteString(LabelStyle.Render("Total:"))
	b.WriteString(ValueStyle.Render(fmt.Sprintf("%d", m.TotalConnections)))
	b.WriteString("    ")

	b.WriteString(LabelStyle.Render("Open:"))
	b.WriteString(ValueStyle.Render(fmt.Sprintf("%d", m.OpenConnections)))
	b.WriteString("\n")

	// Response times
	if m.TotalConnections > 0 {
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
	b.WriteString("\n\n")

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
	return fmt.Sprintf("%-8s  %-6s  %-40s  %s  %s",
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
	if len(path) > 40 {
		path = path[:37] + "..."
	}
	pathStr := rowStyle.Render(fmt.Sprintf("%-40s", path))

	// Status code with color
	statusStyle := GetStatusStyle(req.StatusCode)
	statusStr := statusStyle.Render(fmt.Sprintf("%d", req.StatusCode))

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
