package utils

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/watzon/goshot/background"
)

// ParseHexColor parses a hex color string into a color.Color
// Supports 6-digit (RRGGBB) and 8-digit (RRGGBBAA) hex colors
func ParseHexColor(hex string) (color.Color, error) {
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

// ParseGradientStops takes in a string slice of gradient stops and returns
// a slice of background.GradientStop.
func ParseGradientStops(input []string) ([]background.GradientStop, error) {
	var result []background.GradientStop
	for _, part := range input {
		parts := strings.Split(part, ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid gradient stop format: %s; expected hex color and percentage (e.g., #ff0000;50)", part)
		}

		hexColor := strings.TrimSpace(parts[0])
		positionStr := strings.TrimSpace(parts[1])

		color, err := ParseHexColor(hexColor)
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
