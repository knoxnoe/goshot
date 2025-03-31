package term

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
)

// getPrefix determines the type of ANSI sequence
func getPrefix(seq []byte) string {
	switch {
	case ansi.HasCsiPrefix(seq):
		return "CSI"
	case ansi.HasOscPrefix(seq):
		return "OSC"
	case ansi.HasDcsPrefix(seq):
		return "DCS"
	case ansi.HasApcPrefix(seq):
		return "APC"
	default:
		return ""
	}
}

// ansiColor returns the color for a standard ANSI color code (0-7)
func ansiColor(code int, theme *Theme) color.Color {
	return theme.GetColor(code)
}

// ansiBrightColor returns the color for a bright ANSI color code (8-15)
func ansiBrightColor(code int, theme *Theme) color.Color {
	return theme.GetColor(code + 8) // Bright colors start at index 8
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
