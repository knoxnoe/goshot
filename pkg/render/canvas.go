package render

import (
	"image"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/syntax"
)

// Canvas represents a rendering canvas with all necessary configuration
type Canvas struct {
	chrome        chrome.Chrome
	background    background.Background // Optional: if nil, no background will be applied
	syntaxOptions *syntax.HighlightOptions
	renderConfig  *syntax.RenderConfig
}

// NewCanvas creates a new Canvas instance with default options
func NewCanvas() *Canvas {
	return &Canvas{
		chrome:     chrome.NewWindows11Chrome(),
		background: nil, // No background by default
		syntaxOptions: &syntax.HighlightOptions{
			Style:        "dracula",
			TabWidth:     4,
			ShowLineNums: true,
		},
		renderConfig: syntax.DefaultConfig().SetShowLineNumbers(true),
	}
}

// SetChrome sets the chrome renderer
func (c *Canvas) SetChrome(chrome chrome.Chrome) *Canvas {
	c.chrome = chrome
	return c
}

// SetBackground sets the background renderer
func (c *Canvas) SetBackground(bg background.Background) *Canvas {
	c.background = bg
	return c
}

// SetSyntaxOptions sets the syntax highlighting options
func (c *Canvas) SetSyntaxOptions(opts *syntax.HighlightOptions) *Canvas {
	c.syntaxOptions = opts
	return c
}

// SetRenderConfig sets the syntax render configuration
func (c *Canvas) SetRenderConfig(config *syntax.RenderConfig) *Canvas {
	c.renderConfig = config
	return c
}

// RenderCode takes source code and renders it to an image using the canvas configuration
func (c *Canvas) RenderCode(code string) (image.Image, error) {
	// Highlight the code
	highlighted, err := syntax.Highlight(code, c.syntaxOptions)
	if err != nil {
		return nil, err
	}

	// Render the highlighted code to an image
	content, err := highlighted.RenderToImage(c.renderConfig)
	if err != nil {
		return nil, err
	}

	// Render with chrome first
	content, err = c.chrome.Render(content)
	if err != nil {
		return nil, err
	}

	// Apply background if one is set
	if c.background != nil {
		content = c.background.Render(content)
	}

	return content, nil
}
