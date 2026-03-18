package imageio

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
)

func TestReaderLoadsPNGAsFourChannelBGRImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.png")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	source := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	source.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 40})
	source.SetNRGBA(1, 0, color.NRGBA{R: 1, G: 2, B: 3, A: 255})
	if err := png.Encode(file, source); err != nil {
		file.Close()
		t.Fatalf("encode png: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close png: %v", err)
	}

	loaded, err := NewReader().Load(context.Background(), path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.ID != path {
		t.Fatalf("expected image id %q, got %q", path, loaded.ID)
	}
	if loaded.Width != 2 || loaded.Height != 1 {
		t.Fatalf("unexpected size: %dx%d", loaded.Width, loaded.Height)
	}
	if loaded.Channels != 4 || loaded.BitsPerPixel != 32 || loaded.ByteWidth != 8 {
		t.Fatalf("unexpected metadata: %+v", loaded)
	}
	if loaded.ColorMode != shared.ColorModeBGR {
		t.Fatalf("expected BGR color mode, got %s", loaded.ColorMode)
	}
	if loaded.PixelDensity.ScaleX != 1 || loaded.PixelDensity.ScaleY != 1 {
		t.Fatalf("unexpected density: %+v", loaded.PixelDensity)
	}

	expected := []byte{30, 20, 10, 40, 3, 2, 1, 255}
	if len(loaded.Data) != len(expected) {
		t.Fatalf("expected %d bytes, got %d", len(expected), len(loaded.Data))
	}
	for i, want := range expected {
		if loaded.Data[i] != want {
			t.Fatalf("unexpected byte at %d: got %d want %d", i, loaded.Data[i], want)
		}
	}
}

func TestWriterStoresPNGFromRGBAndBGRImages(t *testing.T) {
	dir := t.TempDir()
	writer := NewWriter()

	rgbPath := filepath.Join(dir, "rgb.png")
	rgbImage := shared.Image{
		Width:     1,
		Height:    1,
		Channels:  4,
		ByteWidth: 4,
		ColorMode: shared.ColorModeRGB,
		Data:      []byte{11, 22, 33, 44},
	}
	if err := writer.Store(context.Background(), provider.ImageWriterParameters{Image: rgbImage, Path: rgbPath}); err != nil {
		t.Fatalf("store rgb image: %v", err)
	}
	assertPNGPixel(t, rgbPath, color.NRGBA{R: 11, G: 22, B: 33, A: 44})

	bgrPath := filepath.Join(dir, "bgr.png")
	bgrImage := shared.Image{
		Width:     1,
		Height:    1,
		Channels:  3,
		ByteWidth: 3,
		ColorMode: shared.ColorModeBGR,
		Data:      []byte{33, 22, 11},
	}
	if err := writer.Store(context.Background(), provider.ImageWriterParameters{Image: bgrImage, Path: bgrPath}); err != nil {
		t.Fatalf("store bgr image: %v", err)
	}
	assertPNGPixel(t, bgrPath, color.NRGBA{R: 11, G: 22, B: 33, A: 255})
}

func TestWriterReturnsStableUnsupportedFormatError(t *testing.T) {
	err := NewWriter().Store(context.Background(), provider.ImageWriterParameters{
		Image: shared.Image{Width: 1, Height: 1, Channels: 4, ByteWidth: 4, ColorMode: shared.ColorModeRGB, Data: []byte{0, 0, 0, 255}},
		Path:  filepath.Join(t.TempDir(), "image.bmp"),
	})
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected unsupported format error, got %v", err)
	}
}

func TestWriterHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewWriter().Store(ctx, provider.ImageWriterParameters{
		Image: shared.Image{Width: 1, Height: 1, Channels: 4, ByteWidth: 4, ColorMode: shared.ColorModeRGB, Data: []byte{0, 0, 0, 255}},
		Path:  filepath.Join(t.TempDir(), "image.png"),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
}

func assertPNGPixel(t *testing.T, path string, want color.NRGBA) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open png: %v", err)
	}
	defer file.Close()

	decoded, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	pixel := color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA)
	if pixel != want {
		t.Fatalf("unexpected pixel: got %#v want %#v", pixel, want)
	}
}
