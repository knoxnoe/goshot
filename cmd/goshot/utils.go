package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/fonts"
)

// Helper functions
func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8

	switch len(hex) {
	case 6:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return nil, err
		}
		a = 255
	case 8:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}

func parseHighlightLines(input string) ([]int, error) {
	var result []int
	parts := strings.Split(input, ";")

	for _, part := range parts {
		if strings.Contains(part, "-") {
			// Handle range (e.g., "1-3")
			var start, end int
			if _, err := fmt.Sscanf(part, "%d-%d", &start, &end); err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			// Handle single line
			var line int
			if _, err := fmt.Sscanf(part, "%d", &line); err != nil {
				return nil, err
			}
			result = append(result, line)
		}
	}

	return result, nil
}

// parseFonts takes in a string of fonts and returns the first font
// that is available on the system.
// Ex. "JetBrains Mono; DejaVu Sans=30"
func parseFonts(input string) (string, float64) {
	for _, fontSpec := range strings.Split(input, ";") {
		parts := strings.Split(strings.TrimSpace(fontSpec), "=")
		fontName := strings.TrimSpace(parts[0])
		fontSize := 14.0
		if len(parts) > 1 {
			if parsedSize, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
				fontSize = parsedSize
			}
		}

		if fonts.IsFontAvailable(fontName) {
			return fontName, fontSize
		}
	}

	return "", 14.0
}

// parseGradientStops takes in a string slice of gradient stops and returns
// a slice of background.GradientStop.
func parseGradientStops(input []string) ([]background.GradientStop, error) {
	var result []background.GradientStop
	for _, part := range input {
		parts := strings.Split(part, ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid gradient stop format: %s; expected hex color and percentage (e.g., #ff0000;50)", part)
		}

		hexColor := strings.TrimSpace(parts[0])
		positionStr := strings.TrimSpace(parts[1])

		color, err := parseHexColor(hexColor)
		if err != nil {
			return nil, fmt.Errorf("invalid color in gradient stop: %s", err)
		}

		position, err := strconv.ParseFloat(positionStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid position in gradient stop: %s", err)
		}

		if position < 0 || position > 100 {
			return nil, fmt.Errorf("gradient stop position must be between 0 and 100: %f", position)
		}

		result = append(result, background.GradientStop{
			Color:    color,
			Position: position / 100, // Convert percentage to decimal
		})
	}
	return result, nil
}
