package fonts

import (
	"testing"
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
			fontName:  "JetBrains",
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
	expectedFonts := []string{"Cantarell", "JetBrains", "Inter"}
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

func TestFontStyle(t *testing.T) {
	style := FontStyle{
		Weight: WeightBold,
		Italic: true,
	}

	font, err := GetFont("Cantarell", &style)
	if err != nil {
		t.Fatalf("Failed to get font with style: %v", err)
	}

	if font.Style.Weight != WeightBold {
		t.Errorf("Font weight = %v, want %v", font.Style.Weight, WeightBold)
	}

	if !font.Style.Italic {
		t.Error("Font should be italic")
	}
}
