package defaults

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PandelisZ/gut/imageproc"
	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
)

var defaultsTestBinarySalt = "gut-defaults-test-salt-1"

type stubWindowProvider struct {
	windows []shared.WindowHandle
	titles  map[shared.WindowHandle]string
}

func (s *stubWindowProvider) GetWindows(context.Context) ([]shared.WindowHandle, error) {
	return append([]shared.WindowHandle(nil), s.windows...), nil
}

func (s *stubWindowProvider) GetActiveWindow(context.Context) (shared.WindowHandle, error) {
	return 0, errors.New("not implemented")
}

func (s *stubWindowProvider) GetWindowTitle(_ context.Context, handle shared.WindowHandle) (string, error) {
	return s.titles[handle], nil
}

func (s *stubWindowProvider) GetWindowRegion(context.Context, shared.WindowHandle) (shared.Region, error) {
	return shared.Region{}, errors.New("not implemented")
}

func (s *stubWindowProvider) FocusWindow(context.Context, shared.WindowHandle) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *stubWindowProvider) MoveWindow(context.Context, shared.WindowHandle, shared.Point) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *stubWindowProvider) ResizeWindow(context.Context, shared.WindowHandle, shared.Size) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *stubWindowProvider) MinimizeWindow(context.Context, shared.WindowHandle) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *stubWindowProvider) RestoreWindow(context.Context, shared.WindowHandle) (bool, error) {
	return false, errors.New("not implemented")
}

func TestRegisterAuxiliaryResolvesExpectedProviders(t *testing.T) {
	registry := provider.NewRegistry()
	windows := &stubWindowProvider{
		windows: []shared.WindowHandle{1, 2},
		titles: map[shared.WindowHandle]string{
			1: "Terminal",
			2: "Editor",
		},
	}
	registry.RegisterWindow(windows)

	RegisterAuxiliary(registry)

	if _, err := registry.Clipboard(); err != nil {
		t.Fatalf("clipboard lookup failed: %v", err)
	}
	if _, err := registry.ImageReader(); err != nil {
		t.Fatalf("image reader lookup failed: %v", err)
	}
	if _, err := registry.ImageWriter(); err != nil {
		t.Fatalf("image writer lookup failed: %v", err)
	}
	if _, err := registry.ImageProcessor(); err != nil {
		t.Fatalf("image processor lookup failed: %v", err)
	}
	if _, err := registry.ColorFinder(); err != nil {
		t.Fatalf("color finder lookup failed: %v", err)
	}
	if _, err := registry.ImageFinder(); err != nil {
		t.Fatalf("image finder lookup failed: %v", err)
	}
	if _, err := registry.TextFinder(); err != nil {
		t.Fatalf("text finder lookup failed: %v", err)
	}

	windowFinder, err := registry.WindowFinder()
	if err != nil {
		t.Fatalf("window finder lookup failed: %v", err)
	}
	match, err := windowFinder.FindMatch(context.Background(), shared.NewWindowQuery("terminal", shared.MatchString("Terminal")))
	if err != nil {
		t.Fatalf("window finder match failed: %v", err)
	}
	if match != 1 {
		t.Fatalf("expected window handle 1, got %d", match)
	}
}

func TestRegisterAuxiliaryRegistersUnavailableWindowFinderWithoutWindowProvider(t *testing.T) {
	registry := provider.NewRegistry()
	RegisterAuxiliary(registry)

	windowFinder, err := registry.WindowFinder()
	if err != nil {
		t.Fatalf("window finder lookup failed: %v", err)
	}
	_, err = windowFinder.FindMatch(context.Background(), shared.NewWindowQuery("terminal", shared.MatchString("Terminal")))
	if !errors.Is(err, imageproc.ErrWindowFinderUnavailable) {
		t.Fatalf("expected unavailable window finder error, got %v", err)
	}
}

func TestRegisterAuxiliaryPNGSmokePath(t *testing.T) {
	registry := provider.NewRegistry()
	RegisterAuxiliary(registry)

	writer, err := registry.ImageWriter()
	if err != nil {
		t.Fatalf("image writer lookup failed: %v", err)
	}
	reader, err := registry.ImageReader()
	if err != nil {
		t.Fatalf("image reader lookup failed: %v", err)
	}
	processor, err := registry.ImageProcessor()
	if err != nil {
		t.Fatalf("image processor lookup failed: %v", err)
	}
	colorFinder, err := registry.ColorFinder()
	if err != nil {
		t.Fatalf("color finder lookup failed: %v", err)
	}

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.png")
	outputPath := filepath.Join(dir, "output.png")
	source := shared.Image{
		Width:     2,
		Height:    2,
		Channels:  4,
		ByteWidth: 8,
		ColorMode: shared.ColorModeRGB,
		Data: []byte{
			1, 2, 3, 255, 10, 20, 30, 255,
			40, 50, 60, 255, 70, 80, 90, 128,
		},
	}

	if err := writer.Store(context.Background(), provider.ImageWriterParameters{Image: source, Path: inputPath}); err != nil {
		t.Fatalf("write input png: %v", err)
	}
	loaded, err := reader.Load(context.Background(), inputPath)
	if err != nil {
		t.Fatalf("read input png: %v", err)
	}
	pixel, err := processor.ColorAt(context.Background(), loaded, shared.Point{X: 1, Y: 1})
	if err != nil {
		t.Fatalf("color at failed: %v", err)
	}
	if pixel != (shared.RGBA{R: 70, G: 80, B: 90, A: 128}) {
		t.Fatalf("unexpected loaded pixel: %#v", pixel)
	}

	match, err := colorFinder.FindMatch(context.Background(), shared.MatchRequest[shared.ColorQuery, any]{
		Haystack: loaded,
		Needle:   shared.NewColorQuery("needle", shared.RGBA{R: 10, G: 20, B: 30, A: 255}),
	})
	if err != nil {
		t.Fatalf("color finder failed: %v", err)
	}
	if match.Location != (shared.Point{X: 1, Y: 0}) {
		t.Fatalf("unexpected match location: %#v", match.Location)
	}

	if err := writer.Store(context.Background(), provider.ImageWriterParameters{Image: loaded, Path: outputPath}); err != nil {
		t.Fatalf("write output png: %v", err)
	}
	reloaded, err := reader.Load(context.Background(), outputPath)
	if err != nil {
		t.Fatalf("read output png: %v", err)
	}
	reloadedPixel, err := processor.ColorAt(context.Background(), reloaded, shared.Point{X: 0, Y: 1})
	if err != nil {
		t.Fatalf("color at reloaded failed: %v", err)
	}
	if reloadedPixel != (shared.RGBA{R: 40, G: 50, B: 60, A: 255}) {
		t.Fatalf("unexpected reloaded pixel: %#v", reloadedPixel)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output png to exist: %v", err)
	}
}
