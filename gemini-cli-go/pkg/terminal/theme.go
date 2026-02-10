package terminal

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Background   lipgloss.Color
	Foreground   lipgloss.Color
	LightBlue    lipgloss.Color
	AccentBlue   lipgloss.Color
	AccentPurple lipgloss.Color
	AccentCyan   lipgloss.Color
	AccentGreen  lipgloss.Color
	AccentYellow lipgloss.Color
	AccentRed    lipgloss.Color
	DiffAdded    lipgloss.Color
	DiffRemoved  lipgloss.Color
	Comment      lipgloss.Color
	Gray         lipgloss.Color
	DarkGray     lipgloss.Color
}

var DefaultDark = Theme{
	Background:   lipgloss.Color("#1E1E2E"),
	Foreground:   lipgloss.Color("#FAFAFA"), // Assuming light foreground for dark mode
	LightBlue:    lipgloss.Color("#ADD8E6"),
	AccentBlue:   lipgloss.Color("#89B4FA"),
	AccentPurple: lipgloss.Color("#CBA6F7"),
	AccentCyan:   lipgloss.Color("#89DCEB"),
	AccentGreen:  lipgloss.Color("#A6E3A1"),
	AccentYellow: lipgloss.Color("#F9E2AF"),
	AccentRed:    lipgloss.Color("#F38BA8"),
	DiffAdded:    lipgloss.Color("#28350B"),
	DiffRemoved:  lipgloss.Color("#430000"),
	Comment:      lipgloss.Color("#6C7086"),
	Gray:         lipgloss.Color("#6C7086"),
	DarkGray:     lipgloss.Color("#45475A"), // Approximate dark gray
}

// Global theme instance, can be updated
var CurrentTheme = DefaultDark

func SetTheme(t Theme) {
	CurrentTheme = t
}

// Helper styles
var (
	StyleForeground   = lipgloss.NewStyle().Foreground(CurrentTheme.Foreground)
	StyleAccentBlue   = lipgloss.NewStyle().Foreground(CurrentTheme.AccentBlue)
	StyleAccentGreen  = lipgloss.NewStyle().Foreground(CurrentTheme.AccentGreen)
	StyleAccentYellow = lipgloss.NewStyle().Foreground(CurrentTheme.AccentYellow)
	StyleAccentRed    = lipgloss.NewStyle().Foreground(CurrentTheme.AccentRed)
	StyleGray         = lipgloss.NewStyle().Foreground(CurrentTheme.Gray)
	StyleComment      = lipgloss.NewStyle().Foreground(CurrentTheme.Comment)
)
