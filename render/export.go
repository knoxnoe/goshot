package render

import (
	"image/jpeg"
	"image/png"
	"os"

	"golang.org/x/image/bmp"
)

// SaveAsPNG saves an image to a file in PNG format
func (c *Canvas) SaveAsPNG(filename string) error {
	img, err := c.RenderToImage()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// SaveAsJPEG saves an image to a file in JPEG format
func (c *Canvas) SaveAsJPEG(filename string) error {
	img, err := c.RenderToImage()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, nil)
}

// SaveAsBMP saves an image to a file in BMP format
func (c *Canvas) SaveAsBMP(filename string) error {
	img, err := c.RenderToImage()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return bmp.Encode(f, img)
}

// SaveAsSVG saves an image to a file in SVG format
func (c *Canvas) SaveAsSVG(filename string) error {
	// TODO
	return nil
}
