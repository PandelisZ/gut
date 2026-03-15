package shared

import "errors"

type PixelDensity struct {
	ScaleX float64
	ScaleY float64
}

type Image struct {
	Width        int
	Height       int
	Data         []byte
	Channels     int
	ID           string
	BitsPerPixel int
	ByteWidth    int
	ColorMode    ColorMode
	PixelDensity PixelDensity
}

func NewImage(width, height int, data []byte, channels int, id string, bitsPerPixel int, byteWidth int, colorMode ColorMode, pixelDensity PixelDensity) (Image, error) {
	image := Image{
		Width:        width,
		Height:       height,
		Data:         data,
		Channels:     channels,
		ID:           id,
		BitsPerPixel: bitsPerPixel,
		ByteWidth:    byteWidth,
		ColorMode:    colorMode,
		PixelDensity: pixelDensity,
	}
	return image, image.Validate()
}

func (i Image) Validate() error {
	if i.Width < 0 || i.Height < 0 {
		return errors.New("image dimensions must be non-negative")
	}
	if i.Channels <= 0 {
		return errors.New("image channels must be greater than zero")
	}
	if i.BitsPerPixel < 0 {
		return errors.New("image bitsPerPixel must be non-negative")
	}
	if i.ByteWidth < 0 {
		return errors.New("image byteWidth must be non-negative")
	}
	if i.PixelDensity.ScaleX == 0 {
		i.PixelDensity.ScaleX = 1
	}
	if i.PixelDensity.ScaleY == 0 {
		i.PixelDensity.ScaleY = 1
	}
	return nil
}

func (i Image) Area() int {
	return i.Width * i.Height
}

func (i Image) HasAlphaChannel() bool {
	return i.Channels > 3
}

func (i Image) NormalizedPixelDensity() PixelDensity {
	density := i.PixelDensity
	if density.ScaleX == 0 {
		density.ScaleX = 1
	}
	if density.ScaleY == 0 {
		density.ScaleY = 1
	}
	return density
}

func IsImage(v any) bool {
	switch v.(type) {
	case Image, *Image:
		return true
	default:
		return false
	}
}
