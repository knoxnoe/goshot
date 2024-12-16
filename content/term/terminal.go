package term

func NewTerminal(style *TermStyle, theme *Theme) *Terminal {
	t := &Terminal{
		Width:         style.Width,
		Height:        style.Height,
		AutoSize:      style.AutoSize,
		PaddingLeft:   style.PaddingLeft,
		PaddingRight:  style.PaddingRight,
		PaddingTop:    style.PaddingTop,
		PaddingBottom: style.PaddingBottom,
		DefaultFg:     theme.GetForeground(),
		DefaultBg:     theme.GetBackground(),
		CurrFg:        theme.GetForeground(),
		CurrBg:        theme.GetBackground(),
		Style:         theme,
		// Initialize cursor position at the start of the content area
		CursorX: style.PaddingLeft,
		CursorY: style.PaddingTop,
	}

	// Initialize cells
	t.Resize(style.Width, style.Height)
	return t
}

func (t *Terminal) Reset() {
	t.CursorX = t.PaddingLeft
	t.CursorY = t.PaddingTop
	t.CurrAttrs = Attributes{}
	t.CurrFg = t.DefaultFg
	t.CurrBg = t.DefaultBg
}

func (t *Terminal) Resize(width, height int) {
	// Add padding to dimensions
	totalWidth := width + t.PaddingLeft + t.PaddingRight
	totalHeight := height + t.PaddingTop + t.PaddingBottom

	newCells := make([][]Cell, totalHeight)
	for i := range newCells {
		newCells[i] = make([]Cell, totalWidth)
		// Initialize with default colors and empty runes
		for j := range newCells[i] {
			newCells[i][j] = Cell{
				Char:    ' ',
				FgColor: t.DefaultFg,
				BgColor: t.DefaultBg,
			}
		}
	}

	// Copy existing content, accounting for padding
	for y := 0; y < min(len(t.Cells), totalHeight); y++ {
		for x := 0; x < min(len(t.Cells[y]), totalWidth); x++ {
			newCells[y][x] = t.Cells[y][x]
		}
	}

	t.Cells = newCells
	t.Width = totalWidth
	t.Height = totalHeight
}

func (t *Terminal) SetCell(x, y int, ch rune) {
	// Ignore any attempts to set cells with negative coordinates
	// or before padding
	if x < t.PaddingLeft || y < t.PaddingTop {
		return
	}

	// If auto-sizing is enabled, track the maximum dimensions regardless of Width/Height
	if t.AutoSize {
		t.MaxX = max(t.MaxX, x+1)
		t.MaxY = max(t.MaxY, y+1)
		// Resize if needed
		if y >= len(t.Cells) || x >= len(t.Cells[0]) {
			t.Resize(max(t.Width, x+1), max(t.Height, y+1))
		}
	} else {
		// If not auto-sizing, respect the terminal boundaries including padding
		if x >= t.Width || y >= t.Height {
			return
		}
	}

	// Ensure we have enough rows
	for len(t.Cells) <= y {
		t.Cells = append(t.Cells, make([]Cell, t.Width))
	}

	// Ensure the row has enough columns
	if len(t.Cells[y]) <= x {
		newRow := make([]Cell, t.Width)
		copy(newRow, t.Cells[y])
		t.Cells[y] = newRow
	}

	t.Cells[y][x] = Cell{
		Char:    ch,
		FgColor: t.CurrFg,
		BgColor: t.CurrBg,
		Attrs:   t.CurrAttrs,
	}
}

func (t *Terminal) NewLine() {
	t.CursorX = t.PaddingLeft
	t.CursorY++
	// If we're at the top padding, skip to the content area
	if t.CursorY < t.PaddingTop {
		t.CursorY = t.PaddingTop
	}
	if t.Height > 0 && t.CursorY >= t.Height {
		// Scroll up, preserving padding area
		copy(t.Cells[t.PaddingTop:], t.Cells[t.PaddingTop+1:])
		t.CursorY = t.Height - 1
		// Clear the new line
		for x := range t.Cells[t.CursorY] {
			t.Cells[t.CursorY][x] = Cell{
				Char:    ' ',
				FgColor: t.DefaultFg,
				BgColor: t.DefaultBg,
			}
		}
	}
}
