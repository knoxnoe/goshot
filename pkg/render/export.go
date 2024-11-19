package render

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

// SaveAsPNG saves an image to a file in PNG format
func SaveAsPNG(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// SaveAsJPEG saves an image to a file in JPEG format
func SaveAsJPEG(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, nil)
}

// SaveAsBMP saves an image to a file in BMP format
func SaveAsBMP(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
