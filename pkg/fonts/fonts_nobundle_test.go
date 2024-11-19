//go:build !bundle_fonts

package fonts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNoBundleFontFS(t *testing.T) {
	fs := FontFS()
	if fs == nil {
		t.Fatal("Expected system FontFS to not be nil")
	}
}

func TestNoBundleGetFont(t *testing.T) {
	// Create a temporary directory for test fonts
	tmpDir, err := os.MkdirTemp("", "goshot-nobundle-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original paths and restore after test
	origPaths := systemFontPaths
	defer func() {
		systemFontPaths = origPaths
	}()

	// Override system font paths for testing
	systemFontPaths = map[string][]string{
		"linux":   {tmpDir},
		"darwin":  {tmpDir},
		"windows": {tmpDir},
	}

	// Create a test font file
	testFontPath := filepath.Join(tmpDir, "TestFont-Regular.ttf")
	if err := os.WriteFile(testFontPath, []byte("mock font data"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		fontName string
		style    *FontStyle
		wantErr  bool
	}{
		{
			name:     "System Font Regular",
			fontName: "TestFont",
			style:    nil,
			wantErr:  false,
		},
		{
			name:     "NonExistent Font",
			fontName: "NonExistentFont",
			style:    nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetFont(tt.fontName, tt.style)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFont() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && font == nil {
				t.Error("GetFont() returned nil font when error not expected")
			}
		})
	}
}
