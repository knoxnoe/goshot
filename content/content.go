package content

import (
	"image"
)

type Content interface {
	Render() (image.Image, error)
}

type LineRange struct {
	Start int
	End   int
}
