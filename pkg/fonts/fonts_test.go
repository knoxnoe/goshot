package fonts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
		{"TestFont-Light.ttf", []byte("mock font data - light")},
		{"TestFont-Medium.ttf", []byte("mock font data - medium")},
		{"NotoSans-Regular.ttf", []byte("mock font data - noto regular")},
		{"NotoSans-Bold.ttf", []byte("mock font data - noto bold")},
	}

	for _, font := range testFonts {
		path := filepath.Join(testFontDir, font.name)
		if err := os.WriteFile(path, font.data, 0644); err != nil {
			panic("failed to create test font file: " + err.Error())
		}
	}
}

func clearFontCache() {
	fontCacheMu.Lock()
	fontCache = make(map[string][]*Font)
	fontCacheMu.Unlock()

	variantCacheMu.Lock()
	variantCache = make(map[string]*Font)
	variantCacheMu.Unlock()
}

func TestGetFont(t *testing.T) {
	clearFontCache()

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
			name:     "Noto Sans Regular",
			fontName: "NotoSans",
			style:    nil,
			wantErr:  false,
			checkFont: func(t *testing.T, f *Font) {
				if f == nil {
					t.Fatal("Expected font, got nil")
				}
				if f.Style.Weight != WeightRegular {
					t.Errorf("Expected regular weight, got %v", f.Style.Weight)
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
	clearFontCache()

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
			name:     "Noto Sans variants",
			fontName: "NotoSans",
			wantErr:  false,
			check: func(t *testing.T, fonts []*Font) {
				if len(fonts) < 2 {
					t.Error("Expected at least two Noto Sans variants")
				}
				for _, f := range fonts {
					if f.Name != "NotoSans" {
						t.Errorf("Expected NotoSans font, got %s", f.Name)
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

func TestFontCache(t *testing.T) {
	clearFontCache()

	// First call should not be cached
	start := time.Now()
	fonts1, err := GetFontVariants("TestFont")
	if err != nil {
		t.Fatalf("GetFontVariants() error = %v", err)
	}
	firstCallDuration := time.Since(start)

	// Second call should be cached and faster
	start = time.Now()
	fonts2, err := GetFontVariants("TestFont")
	if err != nil {
		t.Fatalf("GetFontVariants() error = %v", err)
	}
	secondCallDuration := time.Since(start)

	// Verify cache is working
	if len(fonts1) != len(fonts2) {
		t.Errorf("Cache returned different number of fonts: first=%d, second=%d", len(fonts1), len(fonts2))
	}

	// Second call should be significantly faster
	if secondCallDuration > firstCallDuration/2 {
		t.Errorf("Cached call not faster: first=%v, second=%v", firstCallDuration, secondCallDuration)
	}

	// Test variant cache
	style := &FontStyle{Weight: WeightBold}
	
	// First call to GetFont
	start = time.Now()
	font1, err := GetFont("TestFont", style)
	if err != nil {
		t.Fatalf("GetFont() error = %v", err)
	}
	firstCallDuration = time.Since(start)

	// Second call should use variant cache
	start = time.Now()
	font2, err := GetFont("TestFont", style)
	if err != nil {
		t.Fatalf("GetFont() error = %v", err)
	}
	secondCallDuration = time.Since(start)

	// Verify variant cache is working
	if font1.FilePath != font2.FilePath {
		t.Error("Variant cache returned different fonts")
	}

	// Second call should be faster
	if secondCallDuration > firstCallDuration/2 {
		t.Errorf("Cached variant call not faster: first=%v, second=%v", firstCallDuration, secondCallDuration)
	}
}

func TestListFonts(t *testing.T) {
	clearFontCache()
	
	fonts := ListFonts()
	if len(fonts) == 0 {
		t.Error("Expected at least one font")
	}

	// Check for our test fonts
	expectedFonts := []string{"TestFont", "NotoSans"}
	for _, expected := range expectedFonts {
		found := false
		for _, font := range fonts {
			if font == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find font %s", expected)
		}
	}
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
