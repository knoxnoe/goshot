package fonts

import (
	"testing"

	"golang.org/x/image/math/fixed"
)

func TestGetFont(t *testing.T) {
	tests := []struct {
		name      string
		fontName  string
		style     *FontStyle
		wantError bool
	}{
		{
			name:      "Get default font",
			fontName:  "Cantarell",
			style:     nil,
			wantError: false,
		},
		{
			name:     "Get font with style",
			fontName: "Cantarell",
			style: &FontStyle{
				Weight: WeightBold,
				Italic: true,
			},
			wantError: false,
		},
		{
			name:      "Non-existent font",
			fontName:  "NonExistentFont",
			style:     nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, tt.style)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFont() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && font == nil {
				t.Error("GetFont() returned nil font when error not expected")
			}
		})
	}
}

func TestGetFallback(t *testing.T) {
	tests := []struct {
		name      string
		variant   FallbackVariant
		wantError bool
	}{
		{
			name:      "Sans fallback",
			variant:   FallbackSans,
			wantError: false,
		},
		{
			name:      "Mono fallback",
			variant:   FallbackMono,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFallback(tt.variant)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFallback() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && font == nil {
				t.Error("GetFallback() returned nil font when error not expected")
			}
		})
	}
}

func TestIsFontAvailable(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		want     bool
	}{
		{
			name:     "Available font",
			fontName: "Cantarell",
			want:     true,
		},
		{
			name:     "Available font with weight",
			fontName: "Inter",
			want:     true,
		},
		{
			name:     "Unavailable font",
			fontName: "NonExistentFont",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFontAvailable(tt.fontName); got != tt.want {
				t.Errorf("IsFontAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFontVariants(t *testing.T) {
	tests := []struct {
		name      string
		fontName  string
		wantError bool
	}{
		{
			name:      "Get JetBrains variants",
			fontName:  "JetBrainsMonoNerdFont",
			wantError: false,
		},
		{
			name:      "Non-existent font variants",
			fontName:  "NonExistentFont",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variants, err := GetFontVariants(tt.fontName)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFontVariants() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(variants) == 0 {
				t.Error("GetFontVariants() returned empty variants when some were expected")
			}
		})
	}
}

func TestListFonts(t *testing.T) {
	fonts := ListFonts()
	if len(fonts) == 0 {
		t.Error("ListFonts() returned empty list, expected at least fallback fonts")
	}

	// Check if our embedded fonts are in the list
	expectedFonts := []string{"Cantarell", "JetBrainsMonoNerdFont", "Inter"}
	for _, expected := range expectedFonts {
		found := false
		for _, font := range fonts {
			if font == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListFonts() did not include embedded font %s", expected)
		}
	}
}

func TestMonospaceDetection(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		wantMono bool
	}{
		{
			name:     "JetBrains (should be mono)",
			fontName: "JetBrainsMonoNerdFont",
			wantMono: true,
		},
		{
			name:     "Inter (should not be mono)",
			fontName: "Inter",
			wantMono: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, nil)
			if err != nil {
				t.Fatalf("Failed to get font %s: %v", tt.fontName, err)
			}

			if got := font.IsMono(); got != tt.wantMono {
				t.Errorf("IsMono() = %v, want %v", got, tt.wantMono)
			}
		})
	}
}

func TestGetMaxWidth(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
	}{
		{
			name:     "Get max width for mono font",
			fontName: "JetBrainsMonoNerdFont",
		},
		{
			name:     "Get max width for proportional font",
			fontName: "Inter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, nil)
			if err != nil {
				t.Fatalf("Failed to get font %s: %v", tt.fontName, err)
			}

			width, err := font.GetMaxWidth()
			if err != nil {
				t.Fatalf("GetMaxWidth() error = %v", err)
			}

			if width <= 0 {
				t.Error("GetMaxWidth() returned zero or negative width")
			}

			// Call twice to test caching
			width2, err := font.GetMaxWidth()
			if err != nil {
				t.Fatalf("Second GetMaxWidth() call error = %v", err)
			}

			if width != width2 {
				t.Errorf("Cached width %v != original width %v", width2, width)
			}
		})
	}
}

func TestGetMonoFace(t *testing.T) {
	tests := []struct {
		name      string
		fontName  string
		size      float64
		cellWidth int
	}{
		{
			name:      "Default cell width mono font",
			fontName:  "JetBrainsMonoNerdFont",
			size:      12,
			cellWidth: 0,
		},
		{
			name:      "Custom cell width mono font",
			fontName:  "JetBrainsMonoNerdFont",
			size:      12,
			cellWidth: 20,
		},
		{
			name:      "Default cell width proportional font",
			fontName:  "Inter",
			size:      12,
			cellWidth: 0,
		},
		{
			name:      "Custom cell width proportional font",
			fontName:  "Inter",
			size:      12,
			cellWidth: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, nil)
			if err != nil {
				t.Fatalf("Failed to get font %s: %v", tt.fontName, err)
			}

			face, err := font.GetMonoFace(tt.size, tt.cellWidth)
			if err != nil {
				t.Fatalf("GetMonoFace() error = %v", err)
			}
			defer face.Close()

			// Test that all ASCII printable characters have the same advance width
			var lastAdvance fixed.Int26_6
			for i := 32; i < 127; i++ {
				advance, ok := face.Face.GlyphAdvance(rune(i))
				if !ok {
					continue
				}

				if lastAdvance != 0 && advance != lastAdvance {
					t.Errorf("Glyph %c advance width %d != previous width %d", i, advance, lastAdvance)
				}
				lastAdvance = advance
			}

			// If custom cell width was specified, verify it's being used
			if tt.cellWidth > 0 {
				advance, ok := face.Face.GlyphAdvance('M')
				if !ok {
					t.Fatal("Failed to get advance width for 'M'")
				}

				expectedWidth := fixed.I(tt.cellWidth)
				if advance != expectedWidth {
					t.Errorf("Cell width = %v, want %v", advance, expectedWidth)
				}
			}
		})
	}
}
