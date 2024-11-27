package content

import (
	"image"
)

type Content interface {
	Render() (image.Image, error)
}
