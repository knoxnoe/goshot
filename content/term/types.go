package term

import (
	"image/color"

	"github.com/watzon/goshot/fonts"
)

type TermStyle struct {
	Args          []string                    // Command and arguments
	Theme         string                      // The terminal theme to use
	Font          *fonts.Font                 // The font to use
	FontSize      float64                     // The font size in points
	LineHeight    float64                     // The line height multiplier
	PaddingLeft   int                         // Padding between the code and the left edge
	PaddingRight  int                         // Padding between the code and the right edge
	PaddingTop    int                         // Padding between the code and the top edge
	PaddingBottom int                         // Padding between the code and the bottom edge
	Width         int                         // Terminal width in cells
	Height        int                         // Terminal height in cells
	AutoSize      bool                        // Whether to automatically size the output to the content
	CellSpacing   int                         // Additional horizontal spacing between cells
	ShowPrompt    bool                        // Whether to show a prompt
	PromptFunc    func(command string) string // Template function that returns the prompt text
}

type TermRenderer struct {
	Output []byte
	Style  *TermStyle
	theme  *Theme // Store theme here
}

type Attributes struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Blink         bool
}

type Cell struct {
	Char    rune
	FgColor color.Color
	BgColor color.Color
	Attrs   Attributes
	IsWide  bool // For handling wide characters
}

type Terminal struct {
	Cells         [][]Cell
	Width         int
	Height        int
	CursorX       int
	CursorY       int
	CurrAttrs     Attributes
	CurrFg        color.Color
	CurrBg        color.Color
	Style         *Theme // Theme colors from theme
	MaxX          int    // For dynamic sizing
	MaxY          int    // For dynamic sizing
	DefaultFg     color.Color
	DefaultBg     color.Color
	AutoSize      bool // Whether to automatically size the terminal
	PaddingLeft   int
	PaddingRight  int
	PaddingTop    int
	PaddingBottom int
}
