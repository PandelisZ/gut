package gutmcp

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/PandelisZ/gut/shared"
)

func encodePNG(source shared.Image) ([]byte, error) {
	img, err := toNRGBA(source)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func toNRGBA(source shared.Image) (*image.NRGBA, error) {
	if source.Width < 0 || source.Height < 0 {
		return nil, fmt.Errorf("image dimensions must be non-negative")
	}
	if source.Channels != 3 && source.Channels != 4 {
		return nil, fmt.Errorf("unsupported image channel layout: %d", source.Channels)
	}
	img := image.NewNRGBA(image.Rect(0, 0, source.Width, source.Height))
	stride := source.ByteWidth
	if stride == 0 {
		stride = source.Width * source.Channels
	}

	for y := 0; y < source.Height; y++ {
		for x := 0; x < source.Width; x++ {
			offset := y*stride + x*source.Channels
			if offset < 0 || offset+source.Channels > len(source.Data) {
				return nil, fmt.Errorf("image data too short for point (%d,%d)", x, y)
			}
			alpha := uint8(255)
			if source.Channels == 4 {
				alpha = source.Data[offset+3]
			}
			var pixel color.NRGBA
			switch source.ColorMode {
			case shared.ColorModeRGB:
				pixel = color.NRGBA{
					R: source.Data[offset],
					G: source.Data[offset+1],
					B: source.Data[offset+2],
					A: alpha,
				}
			case shared.ColorModeBGR:
				pixel = color.NRGBA{
					R: source.Data[offset+2],
					G: source.Data[offset+1],
					B: source.Data[offset],
					A: alpha,
				}
			default:
				return nil, fmt.Errorf("unsupported color mode: %s", source.ColorMode)
			}
			img.SetNRGBA(x, y, pixel)
		}
	}
	return img, nil
}
