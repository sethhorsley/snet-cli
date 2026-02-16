package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors - ngrok-inspired purple/blue theme
var (
	ColorPrimary   = lipgloss.Color("#7B61FF") // Purple
	ColorSecondary = lipgloss.Color("#00D9FF") // Cyan
	ColorSuccess   = lipgloss.Color("#00FF87") // Green
	ColorWarning   = lipgloss.Color("#FFD700") // Yellow
	ColorError     = lipgloss.Color("#FF6B6B") // Red
	ColorMuted     = lipgloss.Color("#6C7A89") // Gray
	ColorBorder    = lipgloss.Color("#3E4C59") // Dark gray
	ColorWhite     = lipgloss.Color("#FFFFFF")
)

// Styles
var (
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StatusOnlineStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusOfflineStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	StatusConnectingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	// Info section styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(20)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	URLStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Underline(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	TableRowAltStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// HTTP method styles
	MethodGETStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	MethodPOSTStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	MethodPUTStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	MethodDELETEStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	MethodOtherStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true)

	// Status code styles
	Status2xxStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	Status3xxStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	Status4xxStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	Status5xxStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	// Footer styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	ShortcutKeyStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// Help styles
	HelpTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			Width(15)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// Divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorBorder)
)

// GetMethodStyle returns the style for an HTTP method
func GetMethodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return MethodGETStyle
	case "POST":
		return MethodPOSTStyle
	case "PUT", "PATCH":
		return MethodPUTStyle
	case "DELETE":
		return MethodDELETEStyle
	default:
		return MethodOtherStyle
	}
}

// GetStatusStyle returns the style for an HTTP status code
func GetStatusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return Status2xxStyle
	case code >= 300 && code < 400:
		return Status3xxStyle
	case code >= 400 && code < 500:
		return Status4xxStyle
	case code >= 500:
		return Status5xxStyle
	default:
		return TableRowStyle
	}
}

// GetStatusStyle returns the style for session status
func GetStatusTextStyle(status string) lipgloss.Style {
	switch status {
	case "online":
		return StatusOnlineStyle
	case "offline":
		return StatusOfflineStyle
	case "connecting":
		return StatusConnectingStyle
	default:
		return StatusConnectingStyle
	}
}
