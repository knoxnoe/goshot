package fonts

import (
	"os"
	"path/filepath"
	"testing"
)

// testFontDir is where we store our test font files
var testFontDir string

func TestMain(m *testing.M) {
	// Setup
	var err error
	testFontDir, err = os.MkdirTemp("", "goshot-test-fonts-*")
	if err != nil {
		panic("failed to create test font directory: " + err.Error())
	}

	// Create test font files
	createTestFonts()

	// Override system font paths for testing
	origPaths := systemFontPaths
	systemFontPaths = map[string][]string{
		"linux":   {testFontDir},
		"darwin":  {testFontDir},
		"windows": {testFontDir},
	}

	// Run tests
	code := m.Run()

	// Restore original paths
	systemFontPaths = origPaths

	// Cleanup
	os.RemoveAll(testFontDir)

	os.Exit(code)
}

func createTestFonts() {
	// Create empty font files for testing
	testFonts := []struct {
		name string
		data []byte
	}{
		{"TestFont-Regular.ttf", []byte("mock font data - regular")},
		{"TestFont-Bold.ttf", []byte("mock font data - bold")},
		{"TestFont-Italic.ttf", []byte("mock font data - italic")},
		{"TestFont-BoldItalic.ttf", []byte("mock font data - bold italic")},
	}

	for _, font := range testFonts {
		path := filepath.Join(testFontDir, font.name)
		if err := os.WriteFile(path, font.data, 0644); err != nil {
			panic("failed to create test font file: " + err.Error())
		}
	}
}

func TestGetFont(t *testing.T) {
	tests := []struct {
		name      string
		fontName  string
		style     *FontStyle
		wantErr   bool
		checkFont func(*testing.T, *Font)
	}{
		{
			name:     "Regular style",
			fontName: "TestFont",
			style:    nil, // Should get regular style
			wantErr:  false,
			checkFont: func(t *testing.T, f *Font) {
				if f == nil {
					t.Fatal("Expected font, got nil")
				}
				if f.Name != "TestFont" {
					t.Errorf("Expected TestFont font, got %s", f.Name)
				}
				if f.Style.Weight != WeightRegular {
					t.Errorf("Expected regular weight, got %v", f.Style.Weight)
				}
			},
		},
		{
			name:     "Bold style",
			fontName: "TestFont",
			style: &FontStyle{
				Weight: WeightBold,
			},
			wantErr: false,
			checkFont: func(t *testing.T, f *Font) {
				if f == nil {
					t.Fatal("Expected font, got nil")
				}
				if f.Style.Weight != WeightBold {
					t.Errorf("Expected bold weight, got %v", f.Style.Weight)
				}
			},
		},
		{
			name:     "Nonexistent font",
			fontName: "NonexistentFont",
			style:    nil,
			wantErr:  true,
			checkFont: func(t *testing.T, f *Font) {
				if f != nil {
					t.Error("Expected nil font for nonexistent font")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, tt.style)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFont() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.checkFont(t, font)
		})
	}
}

func TestGetFontVariants(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		wantErr  bool
		check    func(*testing.T, []*Font)
	}{
		{
			name:     "Test font variants",
			fontName: "TestFont",
			wantErr:  false,
			check: func(t *testing.T, fonts []*Font) {
				if len(fonts) == 0 {
					t.Error("Expected at least one font variant")
				}
				for _, f := range fonts {
					if f.Name != "TestFont" {
						t.Errorf("Expected TestFont font, got %s", f.Name)
					}
					if f.Data == nil {
						t.Error("Font data should not be nil")
					}
					if f.FilePath == "" {
						t.Error("Font file path should not be empty")
					}
				}
			},
		},
		{
			name:     "Nonexistent font",
			fontName: "NonexistentFont",
			wantErr:  false, // Should return empty slice, not error
			check: func(t *testing.T, fonts []*Font) {
				if len(fonts) > 0 {
					t.Error("Expected no font variants for nonexistent font")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fonts, err := GetFontVariants(tt.fontName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFontVariants() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.check(t, fonts)
		})
	}
}

func TestListFonts(t *testing.T) {
	fonts := ListFonts()
	if len(fonts) == 0 {
		t.Error("Expected at least one font")
	}

	// Check for our test font
	found := false
	for _, font := range fonts {
		if font == "TestFont" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find TestFont")
	}
}

func TestFontFS(t *testing.T) {
	fs := FontFS()
	if fs == nil {
		t.Fatal("Expected non-nil font filesystem")
	}

	// Try to open the test font directory
	entries, err := fs.Open(".")
	if err != nil {
		t.Fatalf("Failed to open font directory: %v", err)
	}
	defer entries.Close()
}

func TestCleanFontName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Regular font name",
			input:    "Arial",
			expected: "Arial",
		},
		{
			name:     "Font name with weight",
			input:    "Arial-Bold",
			expected: "Arial",
		},
		{
			name:     "Font name with style",
			input:    "Arial-BoldItalic",
			expected: "Arial",
		},
		{
			name:     "Font name with multiple styles",
			input:    "Arial-BoldCondensedItalic",
			expected: "Arial",
		},
		{
			name:     "Font name with version number",
			input:    "Arial1",
			expected: "Arial",
		},
		{
			name:     "Font name with spaces",
			input:    "Times New Roman",
			expected: "Times New Roman",
		},
		{
			name:     "Font name with spaces and style",
			input:    "Times New Roman Bold",
			expected: "Times New Roman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanFontName(tt.input)
			if result != tt.expected {
				t.Errorf("cleanFontName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractFontStyle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     FontStyle
	}{
		{
			name:     "Regular font",
			filename: "Arial-Regular",
			want:     FontStyle{Weight: WeightRegular},
		},
		{
			name:     "Bold font",
			filename: "Arial-Bold",
			want:     FontStyle{Weight: WeightBold},
		},
		{
			name:     "Bold Italic font",
			filename: "Arial-BoldItalic",
			want: FontStyle{
				Weight: WeightBold,
				Italic: true,
			},
		},
		{
			name:     "Light Condensed font",
			filename: "Arial-LightCondensed",
			want: FontStyle{
				Weight:    WeightLight,
				Condensed: true,
			},
		},
		{
			name:     "Mono font",
			filename: "Consolas-Mono",
			want: FontStyle{
				Weight: WeightRegular,
				Mono:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFontStyle(tt.filename)
			if got.Weight != tt.want.Weight {
				t.Errorf("extractFontStyle(%q).Weight = %v, want %v", tt.filename, got.Weight, tt.want.Weight)
			}
			if got.Italic != tt.want.Italic {
				t.Errorf("extractFontStyle(%q).Italic = %v, want %v", tt.filename, got.Italic, tt.want.Italic)
			}
			if got.Condensed != tt.want.Condensed {
				t.Errorf("extractFontStyle(%q).Condensed = %v, want %v", tt.filename, got.Condensed, tt.want.Condensed)
			}
			if got.Mono != tt.want.Mono {
				t.Errorf("extractFontStyle(%q).Mono = %v, want %v", tt.filename, got.Mono, tt.want.Mono)
			}
		})
	}
}
