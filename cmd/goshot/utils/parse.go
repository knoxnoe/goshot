package utils

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/watzon/goshot/pkg/content"
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/fonts"
)

// ParseLineRanges parses line range strings into content.LineRange structs
func ParseLineRanges(input []string) ([]content.LineRange, error) {
	var result []content.LineRange
	for _, part := range input {
		parts := strings.Split(part, "..")
		if len(parts) > 2 {
			return nil, fmt.Errorf("invalid highlight line format: %s; expected start and end line numbers (e.g., ..5)", part)
		}

		if len(parts) == 1 {
			num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid line number: %s", err)
			}
			result = append(result, content.LineRange{Start: num, End: num})
		} else {
			var startNum, endNum int
			var err error

			if parts[0] == "" {
				startNum = 1
			} else {
				startNum, err = strconv.Atoi(strings.TrimSpace(parts[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid start line number: %s", err)
				}
			}

			if parts[1] == "" {
				endNum = -1
			} else {
				endNum, err = strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid end line number: %s", err)
				}
			}

			result = append(result, content.LineRange{Start: startNum, End: endNum})
		}
	}

	return result, nil
}

// ParseFonts takes in a string of fonts and returns the first font
// that is available on the system along with its size.
// Ex. "JetBrains Mono; DejaVu Sans=30"
func ParseFonts(input string) (string, float64) {
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

// ParseRedactionAreas parses redaction area strings into code.RedactionArea structs
func ParseRedactionAreas(areas []string) ([]code.RedactionArea, error) {
	var result []code.RedactionArea
	for _, area := range areas {
		parts := strings.Split(area, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid redaction area format: %s (expected 'x,y,width,height')", area)
		}

		x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid x coordinate in redaction area: %s", area)
		}

		y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid y coordinate in redaction area: %s", area)
		}

		width, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return nil, fmt.Errorf("invalid width in redaction area: %s", area)
		}

		height, err := strconv.Atoi(strings.TrimSpace(parts[3]))
		if err != nil {
			return nil, fmt.Errorf("invalid height in redaction area: %s", area)
		}

		result = append(result, code.RedactionArea{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		})
	}
	return result, nil
}

// DetectLanguage attempts to detect the language from file extension or explicit setting
func DetectLanguage(filename, explicitLanguage string) string {
	// If language is explicitly set, use it
	if explicitLanguage != "" {
		return explicitLanguage
	}

	// If we have a filename, try to detect from extension
	if filename != "" {
		ext := strings.TrimPrefix(filepath.Ext(filename), ".")
		if ext != "" {
			return ext
		}
	}

	// Default to no language
	return ""
}
