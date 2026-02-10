package terminal

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Standard styles using the theme
	SuccessStyle = lipgloss.NewStyle().Foreground(DefaultDark.AccentGreen).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(DefaultDark.AccentRed).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(DefaultDark.AccentYellow).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(DefaultDark.AccentBlue).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(DefaultDark.Gray)
)

func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("✔ " + msg))
}

func PrintError(msg string) {
	fmt.Println(ErrorStyle.Render("✖ " + msg))
}

func PrintWarning(msg string) {
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

func PrintInfo(msg string) {
	fmt.Println(InfoStyle.Render("ℹ " + msg))
}

func PrintMuted(msg string) {
	fmt.Println(MutedStyle.Render(msg))
}

func RenderBox(title, content string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DefaultDark.AccentPurple).
		Padding(1, 2)

	// Create a title style if needed, but for now just render the box
	// Standard lipgloss doesn't have .Title() on Style for borders yet in all versions
	// We can manually add the title

	renderedBox := boxStyle.Render(content)
	if title != "" {
		// Simple manual title (not perfect but works)
		return fmt.Sprintf("%s\n%s", lipgloss.NewStyle().Foreground(DefaultDark.AccentPurple).Bold(true).Render(title), renderedBox)
	}
	return renderedBox
}

func ApplyTheme(t Theme) {
	CurrentTheme = t
	// Re-initialize styles if needed
	SuccessStyle = lipgloss.NewStyle().Foreground(CurrentTheme.AccentGreen).Bold(true)
	ErrorStyle = lipgloss.NewStyle().Foreground(CurrentTheme.AccentRed).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(CurrentTheme.AccentYellow).Bold(true)
	InfoStyle = lipgloss.NewStyle().Foreground(CurrentTheme.AccentBlue).Bold(true)
	MutedStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Gray)
}
