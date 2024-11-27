package term

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/watzon/goshot/pkg/fonts"
)

type TermStyle struct {
	Theme         string      // The chroma syntax theme to use
	Font          *fonts.Font // The font to use
	FontSize      float64     // The font size in points
	LineHeight    float64     // The line height multiplier
	PaddingLeft   int         // Padding between the code and the left edge
	PaddingRight  int         // Padding between the code and the right edge
	PaddingTop    int         // Padding between the code and the top edge
	PaddingBottom int         // Padding between the code and the bottom edge
	MinWidth      int         // Minimum width in pixels (0 means no minimum)
	MaxWidth      int         // Maximum width in pixels (0 means no limit)
}

type TermRenderer struct {
	Output []byte
	Style  *TermStyle
}

func NewRenderer(input []byte, style *TermStyle) *TermRenderer {
	return &TermRenderer{
		Output: input,
		Style:  style,
	}
}

func DefaultRenderer(input []byte) *TermRenderer {
	font, err := fonts.GetFallback(fonts.FallbackMono)
	if err != nil {
		panic(err)
	}

	return NewRenderer(input, &TermStyle{
		Theme:         "monokai",
		Font:          font,
		FontSize:      12,
		LineHeight:    1.2,
		PaddingLeft:   10,
		PaddingRight:  10,
		PaddingTop:    10,
		PaddingBottom: 10,
		MinWidth:      300,
		MaxWidth:      900,
	})
}

func (r *TermRenderer) WithTheme(theme string) *TermRenderer {
	r.Style.Theme = theme
	return r
}

func (r *TermRenderer) WithFont(font *fonts.Font) *TermRenderer {
	r.Style.Font = font
	return r
}

func (r *TermRenderer) WithFontSize(size float64) *TermRenderer {
	r.Style.FontSize = size
	return r
}

func (r *TermRenderer) WithLineHeight(height float64) *TermRenderer {
	r.Style.LineHeight = height
	return r
}

func (r *TermRenderer) WithPadding(left, right, top, bottom int) *TermRenderer {
	r.Style.PaddingLeft = left
	r.Style.PaddingRight = right
	r.Style.PaddingTop = top
	r.Style.PaddingBottom = bottom
	return r
}

func (r *TermRenderer) WithMinWidth(width int) *TermRenderer {
	r.Style.MinWidth = width
	return r
}

func (r *TermRenderer) WithMaxWidth(width int) *TermRenderer {
	r.Style.MaxWidth = width
	return r
}

func (r *TermRenderer) Render() (image.Image, error) {
	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)

	in := r.Output
	for len(in) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(in, state, p)

		s := fmt.Sprintf("%q", seq)
		s = strings.TrimPrefix(s, `"`)
		s = strings.TrimSuffix(s, `"`)

		// Trim introducers and terminators
		// CSI
		s = strings.TrimPrefix(s, "\\x9b")
		s = strings.TrimPrefix(s, "\\x1b[")
		// DCS
		s = strings.TrimPrefix(s, "\\x90")
		s = strings.TrimPrefix(s, "\\x1bP")
		// OSC
		s = strings.TrimPrefix(s, "\\x9d")
		s = strings.TrimPrefix(s, "\\x1b]")
		s = strings.TrimSuffix(s, "\\a")
		// APC
		s = strings.TrimPrefix(s, "\\x9f")
		s = strings.TrimPrefix(s, "\\x1b_")
		// ESC
		if !bytes.Equal(seq, []byte{ansi.ESC}) {
			s = strings.TrimPrefix(s, "\\x1b")
		}
		// ST
		s = strings.TrimSuffix(s, "\\x9c")
		s = strings.TrimSuffix(s, "\\x1b\\\\")

		if width > 0 {
			// This is a character

		} else {
			// This is a control sequence

		}

		in = in[n:]
		state = newState
	}

	return nil, nil
}

func getPrefix(seq []byte) string {
	switch {
	case ansi.HasCsiPrefix(seq):
		return "CSI"
	case ansi.HasOscPrefix(seq):
		return "OSC"
	case ansi.HasDcsPrefix(seq):
		return "DCS"
	case ansi.HasApcPrefix(seq):
		return "APC"
	default:
		return ""
	}
}
