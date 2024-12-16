package code

import (
	"image"
	"image/draw"
	"regexp"
	"sort"

	"github.com/disintegration/imaging"
)

// RedactionPattern represents a pattern to match for redaction
type RedactionPattern struct {
	Pattern *regexp.Regexp
	Name    string
}

// DefaultRedactionPatterns contains common patterns for sensitive information
var DefaultRedactionPatterns = []RedactionPattern{
	// Known secret formats (API keys, tokens, etc)
	{regexp.MustCompile(`` +
		`(?i)` + // Case-insensitive
		`"` + // Opening quote
		`(?P<value>` + // Start value capture
		`(?:` +
		`sk_(?:live|test)_[\w]{24,}|` + // Stripe
		`sk-[\w]{32,}|` + // OpenAI
		`gh[porsu]_[\w]{36,}|` + // GitHub tokens
		`AKIA[\w]{16}|` + // AWS
		`eyJ[\w-_=]+\.eyJ[\w-_=]+\.[\w-_.+/=]+` + // JWT
		`)` +
		`)` + // End value capture
		`"`), "Known Secret Format"},

	// URLs with basic auth - capture password portion
	{regexp.MustCompile(`` +
		`(?i)` + // Case-insensitive
		`(?:\w+)?://` + // Protocol (optional)
		`[^:]+:` + // Username
		`(?P<value>[^@\s]+)` + // Password (captured)
		`@`), "URL Password"},

	// Variables/fields with sensitive names
	{regexp.MustCompile(`` +
		`(?im)` + // Case-insensitive and multiline mode
		`(?:^|\s|"|,)` + // Start of line, whitespace, quote, or comma
		`\s*` + // Optional whitespace
		`(?:")?` + // Optional quote for JSON keys
		`[\w]*` + // Optional prefix
		`(?:` + // Start of sensitive word group
		`key|token|secret|pass(?:word)?|pwd|auth|cred` + // Common sensitive words
		`)` +
		`[\w]*` + // Optional suffix
		`(?:")?\s*` + // Optional quote for JSON keys + whitespace
		`(?:[:=]|:=)\s*` + // Assignment operators (both JSON ":" and code "=")
		`(?:"|\x60)` + // Opening quote or backtick
		`(?P<value>` + // Start value capture
		`(?:[^"\x60]|(?:\n|\\n)[^"\x60]*)*?` + // Content including newlines but not ending quote
		`)` + // End value capture
		`(?:"|\x60)`, // Closing quote or backtick
	), "Sensitive Variable"},
}

// RedactionStyle represents the style of redaction to apply
type RedactionStyle string

const (
	// RedactionStyleBlock replaces text with block characters
	RedactionStyleBlock RedactionStyle = "block"
	// RedactionStyleBlur applies a blur effect to the text
	RedactionStyleBlur RedactionStyle = "blur"
)

// RedactionConfig holds configuration for the redaction feature
type RedactionConfig struct {
	Enabled          bool
	Style            RedactionStyle
	Patterns         []RedactionPattern
	BlurRadius       float64
	ManualRedactions []RedactionArea
}

// RedactionArea represents an area to be redacted in the image
type RedactionArea struct {
	X      int
	Y      int
	Width  int
	Height int
}

// RedactionRange represents a range of text that should be redacted
type RedactionRange struct {
	StartIndex int
	EndIndex   int
	Pattern    string // Name of the pattern that matched
}

// NewRedactionConfig creates a new redaction configuration with default settings
func NewRedactionConfig() *RedactionConfig {
	return &RedactionConfig{
		Enabled:    false,
		Style:      RedactionStyleBlock,
		Patterns:   DefaultRedactionPatterns,
		BlurRadius: 5.0,
	}
}

// AddManualRedaction adds a manual redaction area
func (rc *RedactionConfig) AddManualRedaction(x, y, width, height int) {
	rc.ManualRedactions = append(rc.ManualRedactions, RedactionArea{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	})
}

// FindRedactionRanges analyzes the text and returns ranges that should be redacted
func FindRedactionRanges(config *RedactionConfig, text string) []RedactionRange {
	if !config.Enabled {
		return nil
	}

	var ranges []RedactionRange
	contentBytes := []byte(text)

	// Apply patterns to find matches
	for _, pattern := range config.Patterns {
		matches := pattern.Pattern.FindAllSubmatchIndex(contentBytes, -1)
		for _, match := range matches {
			// The first subgroup is our "value" group that we want to redact
			if len(match) >= 4 { // Make sure we have a subgroup
				ranges = append(ranges, RedactionRange{
					StartIndex: match[2], // Start of first subgroup
					EndIndex:   match[3], // End of first subgroup
					Pattern:    pattern.Name,
				})
			}
		}
	}

	// Sort ranges by start index
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].StartIndex < ranges[j].StartIndex
	})

	// Merge overlapping ranges
	if len(ranges) > 1 {
		merged := make([]RedactionRange, 0, len(ranges))
		current := ranges[0]

		for i := 1; i < len(ranges); i++ {
			if ranges[i].StartIndex <= current.EndIndex {
				// Ranges overlap, merge them
				if ranges[i].EndIndex > current.EndIndex {
					current.EndIndex = ranges[i].EndIndex
				}
				current.Pattern = "merged" // Multiple patterns matched
			} else {
				// No overlap, add current range and start a new one
				merged = append(merged, current)
				current = ranges[i]
			}
		}
		merged = append(merged, current)
		ranges = merged
	}

	return ranges
}

// ShouldRedact returns true if the given position in the text should be redacted
func ShouldRedact(pos int, ranges []RedactionRange) bool {
	for _, r := range ranges {
		if pos >= r.StartIndex && pos < r.EndIndex {
			return true
		}
	}
	return false
}

// redactArea applies a blur effect to a specific area of the image
func redactArea(img *image.RGBA, x, y, width, height int, blurRadius float64) {
	// Create a new RGBA image for the area to be blurred
	bounds := img.Bounds()
	x1 := max(0, x)
	y1 := max(0, y)
	x2 := min(bounds.Max.X, x+width)
	y2 := min(bounds.Max.Y, y+height)

	// Skip if area is outside image bounds
	if x1 >= x2 || y1 >= y2 {
		return
	}

	// Extract the area to be blurred
	area := image.NewRGBA(image.Rect(0, 0, x2-x1, y2-y1))
	draw.Draw(area, area.Bounds(), img, image.Point{x1, y1}, draw.Src)

	// Apply blur
	blurred := imaging.Blur(area, blurRadius)

	// Draw the blurred area back onto the original image
	draw.Draw(img, image.Rect(x1, y1, x2, y2), blurred, image.Point{}, draw.Src)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
