package render

import (
	"fmt"
	"image"

	"github.com/watzon/goshot/pkg/background"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/content"
)

// Canvas represents a rendering canvas with all necessary configuration
type Canvas struct {
	chrome     chrome.Chrome
	background background.Background
	content    content.Content
}

// NewCanvas creates a new Canvas instance with default options
func NewCanvas() *Canvas {
	return &Canvas{
		chrome:     nil,
		background: nil, // No background by default
		content:    nil, // No content by default
	}
}

// WithChrome sets the chrome renderer
func (c *Canvas) WithChrome(chrome chrome.Chrome) *Canvas {
	c.chrome = chrome
	return c
}

// WithBackground sets the background renderer
func (c *Canvas) WithBackground(bg background.Background) *Canvas {
	c.background = bg
	return c
}

// WithContent sets the content renderer
func (c *Canvas) WithContent(content content.Content) *Canvas {
	c.content = content
	return c
}

// RenderToImage renders an image using the given chrome, background, and content;
// all of which are optional, but at least one is required
func (c *Canvas) RenderToImage() (image.Image, error) {
	// Validate that at least one renderer is set
	if c.chrome == nil && c.background == nil && c.content == nil {
		return nil, fmt.Errorf("at least one renderer must be set")
	}

	var img image.Image
	var err error

	// First, render the content
	if c.content != nil {
		img, err = c.content.Render()
		if err != nil {
			return nil, err
		}
	}

	// Then, apply the chrome
	if c.chrome != nil {
		img, err = c.chrome.Render(img)
		if err != nil {
			return nil, err
		}
	}

	// Finally apply the background
	if c.background != nil {
		img, err = c.background.Render(img)
		if err != nil {
			return nil, err
		}
	}

	return img, nil
}
