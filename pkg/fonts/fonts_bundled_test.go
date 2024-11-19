//go:build bundle_fonts

package fonts

import (
	"testing"
)

func TestBundledFontFS(t *testing.T) {
	fs := FontFS()
	if fs == nil {
		t.Fatal("Expected bundled FontFS to not be nil")
	}

	// Test JetBrainsMonoNerdFont is available with various styles
	fonts := []string{
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Regular.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Bold.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Italic.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Light.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Medium.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-ExtraLight.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-SemiBold.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-ExtraBold.ttf",
		"bundled/JetBrainsMonoNerdFont/JetBrainsMonoNerdFont-Thin.ttf",
	}

	for _, fontPath := range fonts {
		f, err := fs.Open(fontPath)
		if err != nil {
			t.Errorf("Expected to find %s in bundled fonts: %v", fontPath, err)
		}
		if f != nil {
			f.Close()
		}
	}
}

func TestBundledGetFont(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		style    *FontStyle
		wantErr  bool
	}{
		{
			name:     "JetBrainsMonoNerdFont Regular",
			fontName: "JetBrainsMonoNerdFont",
			style:    nil,
			wantErr:  false,
		},
		{
			name:     "JetBrainsMonoNerdFont Bold",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightBold,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont Light",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightLight,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont ExtraLight",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightExtraLight,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont Medium",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightMedium,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont SemiBold",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightSemiBold,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont ExtraBold",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Weight: WeightExtraBold,
			},
			wantErr: false,
		},
		{
			name:     "JetBrainsMonoNerdFont Italic",
			fontName: "JetBrainsMonoNerdFont",
			style: &FontStyle{
				Italic: true,
			},
			wantErr: false,
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
			if !tt.wantErr {
				if font == nil {
					t.Error("GetFont() returned nil font when error not expected")
				} else {
					// Verify the font name and style match what we requested
					if font.Name != tt.fontName {
						t.Errorf("GetFont() returned font name %s, want %s", font.Name, tt.fontName)
					}
					if tt.style != nil {
						if font.Style.Weight != tt.style.Weight {
							t.Errorf("GetFont() returned font weight %v, want %v", font.Style.Weight, tt.style.Weight)
						}
						if font.Style.Italic != tt.style.Italic {
							t.Errorf("GetFont() returned font italic %v, want %v", font.Style.Italic, tt.style.Italic)
						}
					}
				}
			}
		})
	}
}
