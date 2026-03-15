package imageio

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"

	"gut/provider"
	"gut/shared"
)

var ErrUnsupportedFormat = errors.New("image format unsupported")

type Reader struct{}

type Writer struct{}

func NewReader() *Reader {
	return &Reader{}
}

func NewWriter() *Writer {
	return &Writer{}
}

func (r *Reader) Load(ctx context.Context, path string) (shared.Image, error) {
	if err := ctx.Err(); err != nil {
		return shared.Image{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return shared.Image{}, err
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		return shared.Image{}, err
	}

	bounds := decoded.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	data := make([]byte, width*height*4)
	stride := width * 4
	for y := 0; y < height; y++ {
		rowOffset := y * stride
		for x := 0; x < width; x++ {
			pixel := color.NRGBAModel.Convert(decoded.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			offset := rowOffset + x*4
			data[offset] = pixel.B
			data[offset+1] = pixel.G
			data[offset+2] = pixel.R
			data[offset+3] = pixel.A
		}
	}

	return shared.Image{
		Width:        width,
		Height:       height,
		Data:         data,
		Channels:     4,
		ID:           path,
		BitsPerPixel: 32,
		ByteWidth:    stride,
		ColorMode:    shared.ColorModeBGR,
		PixelDensity: shared.PixelDensity{ScaleX: 1, ScaleY: 1},
	}, nil
}

func (w *Writer) Store(ctx context.Context, parameters provider.ImageWriterParameters) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Ext(parameters.Path), ".png") {
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, parameters.Path)
	}

	img, err := encodeImage(parameters.Image)
	if err != nil {
		return err
	}

	file, err := os.Create(parameters.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func encodeImage(source shared.Image) (*image.NRGBA, error) {
	stride := source.ByteWidth
	if stride == 0 {
		stride = source.Width * source.Channels
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, source.Width, source.Height))
	for y := 0; y < source.Height; y++ {
		rowOffset := y * stride
		for x := 0; x < source.Width; x++ {
			offset := rowOffset + x*source.Channels
			pixel, err := decodePixel(source.Data, offset, source.Channels, source.ColorMode)
			if err != nil {
				return nil, err
			}
			canvas.SetNRGBA(x, y, pixel)
		}
	}
	return canvas, nil
}

func decodePixel(data []byte, offset int, channels int, colorMode shared.ColorMode) (color.NRGBA, error) {
	if offset < 0 || offset+channels > len(data) {
		return color.NRGBA{}, fmt.Errorf("image data too short for channels=%d at offset=%d", channels, offset)
	}

	var pixel color.NRGBA
	pixel.A = 255
	if channels == 4 {
		pixel.A = data[offset+3]
	} else if channels != 3 {
		return color.NRGBA{}, fmt.Errorf("unsupported image channels: %d", channels)
	}

	switch colorMode {
	case shared.ColorModeRGB:
		pixel.R = data[offset]
		pixel.G = data[offset+1]
		pixel.B = data[offset+2]
	case shared.ColorModeBGR:
		pixel.B = data[offset]
		pixel.G = data[offset+1]
		pixel.R = data[offset+2]
	default:
		return color.NRGBA{}, fmt.Errorf("unsupported color mode: %s", colorMode)
	}
	return pixel, nil
}

var _ provider.ImageReader = (*Reader)(nil)
var _ provider.ImageWriter = (*Writer)(nil)
