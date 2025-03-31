package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Styles defines our CLI styles
var Styles = struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Error        lipgloss.Style
	Info         lipgloss.Style
	GroupTitle   lipgloss.Style
	SuccessBox   lipgloss.Style
	InfoBox      lipgloss.Style
	FlagStyle    lipgloss.Style
	DescStyle    lipgloss.Style
	DefaultStyle lipgloss.Style
}{
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#d7008f", Dark: "#FF79C6"}).
		MarginBottom(1),
	Subtitle: lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#7e3ff2", Dark: "#BD93F9"}).
		MarginBottom(1),
	Error: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#d70000", Dark: "#FF5555"}),
	SuccessBox: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#388E3C"}),
	InfoBox: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(lipgloss.AdaptiveColor{Light: "#0087af", Dark: "#8BE9FD"}),
	Info: lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#0087af", Dark: "#8BE9FD"}),
	GroupTitle: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#d75f00", Dark: "#FFB86C"}),
	FlagStyle:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#50FA7B"}),
	DescStyle:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#303030", Dark: "#F8F8F2"}),
	DefaultStyle: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#6272A4"}).Italic(true),
}

// LogMessage prints a styled message with consistent alignment
func LogMessage(box lipgloss.Style, tag string, message string) {
	// Set a consistent width for the tag box and center the text
	const boxWidth = 11 // 9 characters + 2 padding spaces
	paddedTag := fmt.Sprintf("%*s", -boxWidth, tag)
	centeredBox := box.Width(boxWidth).Align(lipgloss.Center)
	fmt.Fprintln(os.Stderr, centeredBox.Render(paddedTag)+" "+Styles.Info.Render(message))
}
