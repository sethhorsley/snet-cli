package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme
type Theme struct {
	Name             string
	Primary          lipgloss.Color
	Secondary        lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Error            lipgloss.Color
	Muted            lipgloss.Color
	Border           lipgloss.Color
	Text             lipgloss.Color
	Background       lipgloss.Color
	UseTransparentBg bool
}

// Available themes
var themes = map[string]Theme{
	"default": {
		Name:             "Default (Transparent)",
		Primary:          lipgloss.Color("#7B61FF"), // Purple
		Secondary:        lipgloss.Color("#00D9FF"), // Cyan
		Success:          lipgloss.Color("#00FF87"), // Green
		Warning:          lipgloss.Color("#FFD700"), // Yellow
		Error:            lipgloss.Color("#FF6B6B"), // Red
		Muted:            lipgloss.Color("#6C7A89"), // Gray
		Border:           lipgloss.Color("#3E4C59"), // Dark gray
		Text:             lipgloss.Color("#FFFFFF"), // White
		Background:       lipgloss.Color("#1a1a1a"), // Dark (for overlay)
		UseTransparentBg: true,
	},
	"catppuccin-latte": {
		Name:             "Catppuccin Latte",
		Primary:          lipgloss.Color("#8839ef"), // Mauve
		Secondary:        lipgloss.Color("#1e66f5"), // Blue
		Success:          lipgloss.Color("#40a02b"), // Green
		Warning:          lipgloss.Color("#df8e1d"), // Yellow
		Error:            lipgloss.Color("#d20f39"), // Red
		Muted:            lipgloss.Color("#9ca0b0"), // Overlay0
		Border:           lipgloss.Color("#ccd0da"), // Surface1
		Text:             lipgloss.Color("#4c4f69"), // Text
		Background:       lipgloss.Color("#eff1f5"), // Base
		UseTransparentBg: false,
	},
	"catppuccin-frappe": {
		Name:             "Catppuccin Frappé",
		Primary:          lipgloss.Color("#ca9ee6"), // Mauve
		Secondary:        lipgloss.Color("#8caaee"), // Blue
		Success:          lipgloss.Color("#a6d189"), // Green
		Warning:          lipgloss.Color("#e5c890"), // Yellow
		Error:            lipgloss.Color("#e78284"), // Red
		Muted:            lipgloss.Color("#737994"), // Overlay0
		Border:           lipgloss.Color("#414559"), // Surface1
		Text:             lipgloss.Color("#c6d0f5"), // Text
		Background:       lipgloss.Color("#303446"), // Base
		UseTransparentBg: false,
	},
	"catppuccin-macchiato": {
		Name:             "Catppuccin Macchiato",
		Primary:          lipgloss.Color("#c6a0f6"), // Mauve
		Secondary:        lipgloss.Color("#8aadf4"), // Blue
		Success:          lipgloss.Color("#a6da95"), // Green
		Warning:          lipgloss.Color("#eed49f"), // Yellow
		Error:            lipgloss.Color("#ed8796"), // Red
		Muted:            lipgloss.Color("#6e738d"), // Overlay0
		Border:           lipgloss.Color("#363a4f"), // Surface1
		Text:             lipgloss.Color("#cad3f5"), // Text
		Background:       lipgloss.Color("#24273a"), // Base
		UseTransparentBg: false,
	},
	"nord": {
		Name:             "Nord",
		Primary:          lipgloss.Color("#88c0d0"), // Frost 2
		Secondary:        lipgloss.Color("#81a1c1"), // Frost 1
		Success:          lipgloss.Color("#a3be8c"), // Green
		Warning:          lipgloss.Color("#ebcb8b"), // Yellow
		Error:            lipgloss.Color("#bf616a"), // Red
		Muted:            lipgloss.Color("#4c566a"), // Polar Night 3
		Border:           lipgloss.Color("#3b4252"), // Polar Night 1
		Text:             lipgloss.Color("#eceff4"), // Snow Storm 3
		Background:       lipgloss.Color("#2e3440"), // Polar Night 0
		UseTransparentBg: false,
	},
	"dracula": {
		Name:             "Dracula",
		Primary:          lipgloss.Color("#bd93f9"), // Purple
		Secondary:        lipgloss.Color("#8be9fd"), // Cyan
		Success:          lipgloss.Color("#50fa7b"), // Green
		Warning:          lipgloss.Color("#f1fa8c"), // Yellow
		Error:            lipgloss.Color("#ff5555"), // Red
		Muted:            lipgloss.Color("#6272a4"), // Comment
		Border:           lipgloss.Color("#44475a"), // Current Line
		Text:             lipgloss.Color("#f8f8f2"), // Foreground
		Background:       lipgloss.Color("#282a36"), // Background
		UseTransparentBg: false,
	},
	"gruvbox-dark": {
		Name:             "Gruvbox Dark",
		Primary:          lipgloss.Color("#d3869b"), // Purple
		Secondary:        lipgloss.Color("#83a598"), // Blue
		Success:          lipgloss.Color("#b8bb26"), // Green
		Warning:          lipgloss.Color("#fabd2f"), // Yellow
		Error:            lipgloss.Color("#fb4934"), // Red
		Muted:            lipgloss.Color("#928374"), // Gray
		Border:           lipgloss.Color("#504945"), // Dark3
		Text:             lipgloss.Color("#ebdbb2"), // FG
		Background:       lipgloss.Color("#282828"), // BG
		UseTransparentBg: false,
	},
	"tokyo-night": {
		Name:             "Tokyo Night",
		Primary:          lipgloss.Color("#bb9af7"), // Purple
		Secondary:        lipgloss.Color("#7aa2f7"), // Blue
		Success:          lipgloss.Color("#9ece6a"), // Green
		Warning:          lipgloss.Color("#e0af68"), // Yellow
		Error:            lipgloss.Color("#f7768e"), // Red
		Muted:            lipgloss.Color("#565f89"), // Comment
		Border:           lipgloss.Color("#292e42"), // Border
		Text:             lipgloss.Color("#c0caf5"), // FG
		Background:       lipgloss.Color("#1a1b26"), // BG
		UseTransparentBg: false,
	},
	"solarized-dark": {
		Name:             "Solarized Dark",
		Primary:          lipgloss.Color("#d33682"), // Magenta
		Secondary:        lipgloss.Color("#268bd2"), // Blue
		Success:          lipgloss.Color("#859900"), // Green
		Warning:          lipgloss.Color("#b58900"), // Yellow
		Error:            lipgloss.Color("#dc322f"), // Red
		Muted:            lipgloss.Color("#586e75"), // Base01
		Border:           lipgloss.Color("#073642"), // Base02
		Text:             lipgloss.Color("#839496"), // Base0
		Background:       lipgloss.Color("#002b36"), // Base03
		UseTransparentBg: false,
	},
	"catppuccin-mocha": {
		Name:             "Catppuccin Mocha",
		Primary:          lipgloss.Color("#cba6f7"), // Mauve
		Secondary:        lipgloss.Color("#89b4fa"), // Blue
		Success:          lipgloss.Color("#a6e3a1"), // Green
		Warning:          lipgloss.Color("#f9e2af"), // Yellow
		Error:            lipgloss.Color("#f38ba8"), // Red
		Muted:            lipgloss.Color("#6c7086"), // Overlay0
		Border:           lipgloss.Color("#45475a"), // Surface1
		Text:             lipgloss.Color("#cdd6f4"), // Text
		Background:       lipgloss.Color("#1e1e2e"), // Base
		UseTransparentBg: false,
	},
}

// ThemeNames returns all available theme names in order
func ThemeNames() []string {
	return []string{
		"default",
		"catppuccin-mocha",
		"catppuccin-latte",
		"catppuccin-frappe",
		"catppuccin-macchiato",
		"nord",
		"dracula",
		"gruvbox-dark",
		"tokyo-night",
		"solarized-dark",
	}
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	if theme, ok := themes[name]; ok {
		return theme
	}
	// Default to default theme if not found
	return themes["default"]
}

// ApplyTheme applies a theme to all styles
func ApplyTheme(themeName string) {
	theme := GetTheme(themeName)

	// Update colors
	ColorPrimary = theme.Primary
	ColorSecondary = theme.Secondary
	ColorSuccess = theme.Success
	ColorWarning = theme.Warning
	ColorError = theme.Error
	ColorMuted = theme.Muted
	ColorBorder = theme.Border
	ColorWhite = theme.Text
	ColorBackground = theme.Background
	UseTransparentBg = theme.UseTransparentBg

	// Update all styles
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

	LabelStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Width(20)

	ValueStyle = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	URLStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Underline(true)

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

	Status2xxStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	Status3xxStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	Status4xxStyle = lipgloss.NewStyle().
		Foreground(ColorWarning)

	Status5xxStyle = lipgloss.NewStyle().
		Foreground(ColorError)

	FooterStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)

	ShortcutKeyStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	ShortcutDescStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

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

	BoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)

	DividerStyle = lipgloss.NewStyle().
		Foreground(ColorBorder)

	// Theme selector styles
	ThemeSelectorTitleStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		MarginBottom(1)

	ThemeSelectorItemStyle = lipgloss.NewStyle().
		Foreground(ColorWhite).
		PaddingLeft(2)

	ThemeSelectorSelectedStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		PaddingLeft(0)

	ThemeSelectorBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)
}
