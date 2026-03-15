package screen

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gut/provider"
	"gut/shared"
	"gut/window"
)

type fakeScreenProvider struct {
	size               shared.Region
	fullImage          shared.Image
	regionImage        shared.Image
	grabbedRegions     []shared.Region
	highlightedRegions []shared.Region
	highlightDurations []time.Duration
	highlightOpacities []float64
}

func (f *fakeScreenProvider) GrabScreen(context.Context) (shared.Image, error) {
	return f.fullImage, nil
}

func (f *fakeScreenProvider) GrabScreenRegion(_ context.Context, region shared.Region) (shared.Image, error) {
	f.grabbedRegions = append(f.grabbedRegions, region)
	if f.regionImage.Width != 0 || f.regionImage.Height != 0 || len(f.regionImage.Data) != 0 {
		return f.regionImage, nil
	}
	return f.fullImage, nil
}

func (f *fakeScreenProvider) HighlightScreenRegion(_ context.Context, region shared.Region, duration time.Duration, opacity float64) error {
	f.highlightedRegions = append(f.highlightedRegions, region)
	f.highlightDurations = append(f.highlightDurations, duration)
	f.highlightOpacities = append(f.highlightOpacities, opacity)
	return nil
}

func (f *fakeScreenProvider) ScreenWidth(context.Context) (int, error) {
	return f.size.Width, nil
}

func (f *fakeScreenProvider) ScreenHeight(context.Context) (int, error) {
	return f.size.Height, nil
}

func (f *fakeScreenProvider) ScreenSize(context.Context) (shared.Region, error) {
	return f.size, nil
}

type fakeColorFinder struct {
	findSequence []colorFindResult
	allMatches   []shared.MatchResult[shared.Point]
	matchErr     error
	findCalls    int
	lastRequest  shared.MatchRequest[shared.ColorQuery, any]
}

type colorFindResult struct {
	match shared.MatchResult[shared.Point]
	err   error
}

func (f *fakeColorFinder) FindMatch(_ context.Context, request shared.MatchRequest[shared.ColorQuery, any]) (shared.MatchResult[shared.Point], error) {
	f.lastRequest = request
	f.findCalls++
	if f.findCalls <= len(f.findSequence) {
		result := f.findSequence[f.findCalls-1]
		return result.match, result.err
	}
	if f.matchErr != nil {
		return shared.MatchResult[shared.Point]{Error: f.matchErr}, f.matchErr
	}
	if len(f.allMatches) == 0 {
		return shared.MatchResult[shared.Point]{Error: errors.New("color not found")}, errors.New("color not found")
	}
	return f.allMatches[0], nil
}

func (f *fakeColorFinder) FindMatches(_ context.Context, request shared.MatchRequest[shared.ColorQuery, any]) ([]shared.MatchResult[shared.Point], error) {
	f.lastRequest = request
	if f.matchErr != nil {
		return nil, f.matchErr
	}
	return append([]shared.MatchResult[shared.Point](nil), f.allMatches...), nil
}

type fakeImageFinder struct {
	match    shared.MatchResult[shared.Region]
	matches  []shared.MatchResult[shared.Region]
	matchErr error
	lastReq  shared.MatchRequest[shared.Image, any]
}

func (f *fakeImageFinder) FindMatch(_ context.Context, request shared.MatchRequest[shared.Image, any]) (shared.MatchResult[shared.Region], error) {
	f.lastReq = request
	if f.matchErr != nil {
		return shared.MatchResult[shared.Region]{Error: f.matchErr}, f.matchErr
	}
	return f.match, nil
}

func (f *fakeImageFinder) FindMatches(_ context.Context, request shared.MatchRequest[shared.Image, any]) ([]shared.MatchResult[shared.Region], error) {
	f.lastReq = request
	if f.matchErr != nil {
		return nil, f.matchErr
	}
	return append([]shared.MatchResult[shared.Region](nil), f.matches...), nil
}

type fakeWindowFinder struct {
	handles []shared.WindowHandle
	err     error
}

func (f *fakeWindowFinder) FindMatch(context.Context, shared.WindowQuery) (shared.WindowHandle, error) {
	if f.err != nil {
		return 0, f.err
	}
	if len(f.handles) == 0 {
		return 0, errMatchNotFound
	}
	return f.handles[0], nil
}

func (f *fakeWindowFinder) FindMatches(context.Context, shared.WindowQuery) ([]shared.WindowHandle, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]shared.WindowHandle(nil), f.handles...), nil
}

type fakeWindowProvider struct {
	windows []shared.WindowHandle
	active  shared.WindowHandle
	titles  map[shared.WindowHandle]string
	regions map[shared.WindowHandle]shared.Region
}

func (f *fakeWindowProvider) GetWindows(context.Context) ([]shared.WindowHandle, error) {
	return append([]shared.WindowHandle(nil), f.windows...), nil
}

func (f *fakeWindowProvider) GetActiveWindow(context.Context) (shared.WindowHandle, error) {
	return f.active, nil
}

func (f *fakeWindowProvider) GetWindowTitle(_ context.Context, handle shared.WindowHandle) (string, error) {
	return f.titles[handle], nil
}

func (f *fakeWindowProvider) GetWindowRegion(_ context.Context, handle shared.WindowHandle) (shared.Region, error) {
	return f.regions[handle], nil
}

func (f *fakeWindowProvider) FocusWindow(context.Context, shared.WindowHandle) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) MoveWindow(context.Context, shared.WindowHandle, shared.Point) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) ResizeWindow(context.Context, shared.WindowHandle, shared.Size) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) MinimizeWindow(context.Context, shared.WindowHandle) (bool, error) {
	return true, nil
}

func (f *fakeWindowProvider) RestoreWindow(context.Context, shared.WindowHandle) (bool, error) {
	return true, nil
}

type fakeImageWriter struct {
	stored []provider.ImageWriterParameters
}

func (f *fakeImageWriter) Store(_ context.Context, parameters provider.ImageWriterParameters) error {
	f.stored = append(f.stored, parameters)
	return nil
}

type fakeImageProcessor struct {
	color     shared.RGBA
	lastPoint shared.Point
}

func (f *fakeImageProcessor) ColorAt(_ context.Context, _ shared.Image, point shared.Point) (shared.RGBA, error) {
	f.lastPoint = point
	return f.color, nil
}

func TestScreenFindAndFindAllColorOffsetResults(t *testing.T) {
	registry := provider.NewRegistry()
	screenProvider := &fakeScreenProvider{
		size:        shared.Region{Left: 0, Top: 0, Width: 40, Height: 30},
		regionImage: shared.Image{Width: 10, Height: 10, Channels: 4, ByteWidth: 40, ColorMode: shared.ColorModeRGB, Data: []byte{0}},
	}
	colorFinder := &fakeColorFinder{
		allMatches: []shared.MatchResult[shared.Point]{
			{Location: shared.Point{X: 1, Y: 2}, Confidence: 1},
			{Location: shared.Point{X: 3, Y: 4}, Confidence: 1},
		},
	}
	registry.RegisterScreen(screenProvider)
	registry.RegisterColorFinder(colorFinder)
	scr := New(registry)

	result, err := scr.Find(context.Background(), shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 4}), &FindOptions{
		SearchRegion: &shared.Region{Left: 5, Top: 7, Width: 10, Height: 10},
	})
	if err != nil {
		t.Fatalf("unexpected color find error: %v", err)
	}
	point, ok := result.(shared.Point)
	if !ok {
		t.Fatalf("unexpected color find result type: %T", result)
	}
	if point != (shared.Point{X: 6, Y: 9}) {
		t.Fatalf("unexpected offset point: %#v", point)
	}
	if colorFinder.lastRequest.Confidence == nil || *colorFinder.lastRequest.Confidence != 1 {
		t.Fatalf("unexpected color confidence: %#v", colorFinder.lastRequest.Confidence)
	}
	if len(screenProvider.grabbedRegions) != 1 || screenProvider.grabbedRegions[0] != (shared.Region{Left: 5, Top: 7, Width: 10, Height: 10}) {
		t.Fatalf("unexpected grabbed regions: %#v", screenProvider.grabbedRegions)
	}

	results, err := scr.FindAll(context.Background(), shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 4}), &FindOptions{
		SearchRegion: &shared.Region{Left: 5, Top: 7, Width: 10, Height: 10},
	})
	if err != nil {
		t.Fatalf("unexpected color find all error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("unexpected color results: %#v", results)
	}
	if results[0].(shared.Point) != (shared.Point{X: 6, Y: 9}) || results[1].(shared.Point) != (shared.Point{X: 8, Y: 11}) {
		t.Fatalf("unexpected offset color results: %#v", results)
	}
}

func TestScreenFindWindowWrapsAndFiltersBySearchRegion(t *testing.T) {
	registry := provider.NewRegistry()
	registry.RegisterScreen(&fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 80, Height: 60}})
	registry.RegisterWindowFinder(&fakeWindowFinder{handles: []shared.WindowHandle{1, 2}})
	registry.RegisterWindow(&fakeWindowProvider{
		regions: map[shared.WindowHandle]shared.Region{
			1: {Left: 4, Top: 5, Width: 10, Height: 10},
			2: {Left: 40, Top: 40, Width: 10, Height: 10},
		},
		titles: map[shared.WindowHandle]string{1: "Editor", 2: "Terminal"},
	})
	scr := New(registry)
	query := shared.NewWindowQuery("editor", shared.MatchString("Editor"))

	found, err := scr.Find(context.Background(), query, &FindOptions{
		SearchRegion: &shared.Region{Left: 0, Top: 0, Width: 20, Height: 20},
	})
	if err != nil {
		t.Fatalf("unexpected window find error: %v", err)
	}
	first, ok := found.(*window.Window)
	if !ok || first.Handle != 1 {
		t.Fatalf("unexpected wrapped window from find: %#v", found)
	}

	results, err := scr.FindAll(context.Background(), query, &FindOptions{
		SearchRegion: &shared.Region{Left: 0, Top: 0, Width: 20, Height: 20},
	})
	if err != nil {
		t.Fatalf("unexpected window find all error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected filtered windows: %#v", results)
	}
	wrapped, ok := results[0].(*window.Window)
	if !ok || wrapped.Handle != 1 {
		t.Fatalf("unexpected wrapped window: %#v", results[0])
	}
}

func TestScreenImageSearchOffsetsRegionsAndAutoHighlights(t *testing.T) {
	registry := provider.NewRegistry()
	screenProvider := &fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 100, Height: 80}, regionImage: shared.Image{Width: 20, Height: 20, Channels: 4, ByteWidth: 80, ColorMode: shared.ColorModeRGB, Data: []byte{0}}}
	imageFinder := &fakeImageFinder{match: shared.MatchResult[shared.Region]{Location: shared.Region{Left: 2, Top: 3, Width: 4, Height: 5}}}
	registry.RegisterScreen(screenProvider)
	registry.RegisterImageFinder(imageFinder)
	scr := New(registry)
	scr.Config.AutoHighlight = true
	scr.Config.HighlightDuration = 25 * time.Millisecond
	scr.Config.HighlightOpacity = 0.75

	result, err := scr.Find(context.Background(), shared.Image{Width: 1, Height: 1, Channels: 4, ByteWidth: 4, ColorMode: shared.ColorModeRGB, Data: []byte{1, 2, 3, 4}}, &FindOptions{
		SearchRegion: &shared.Region{Left: 10, Top: 20, Width: 20, Height: 20},
	})
	if err != nil {
		t.Fatalf("unexpected image find error: %v", err)
	}
	region, ok := result.(shared.Region)
	if !ok {
		t.Fatalf("unexpected image find result type: %T", result)
	}
	if region != (shared.Region{Left: 12, Top: 23, Width: 4, Height: 5}) {
		t.Fatalf("unexpected offset region: %#v", region)
	}
	if len(screenProvider.highlightedRegions) != 1 || screenProvider.highlightedRegions[0] != region {
		t.Fatalf("unexpected highlights: %#v", screenProvider.highlightedRegions)
	}
	if imageFinder.lastReq.Confidence == nil || *imageFinder.lastReq.Confidence != 0.99 {
		t.Fatalf("unexpected default confidence: %#v", imageFinder.lastReq.Confidence)
	}
}

func TestScreenFindImageReturnsMissingProviderErrorWhenUnavailable(t *testing.T) {
	registry := provider.NewRegistry()
	registry.RegisterScreen(&fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 20, Height: 20}})
	scr := New(registry)

	_, err := scr.Find(context.Background(), shared.Image{Width: 1, Height: 1, Channels: 4, ByteWidth: 4, ColorMode: shared.ColorModeRGB, Data: []byte{1, 2, 3, 4}}, nil)
	if !errors.Is(err, provider.ErrMissingImageFinderProvider) {
		t.Fatalf("expected missing image finder error, got %v", err)
	}
}

func TestScreenNewWithNilRegistryReturnsStableMissingProviderError(t *testing.T) {
	_, err := New(nil).Width(context.Background())
	if !errors.Is(err, provider.ErrMissingScreenProvider) {
		t.Fatalf("expected missing screen provider error, got %v", err)
	}
}

func TestScreenWaitForReturnsSuccessAndTimeoutIncludesLastError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		registry := provider.NewRegistry()
		registry.RegisterScreen(&fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 20, Height: 20}, regionImage: shared.Image{Width: 10, Height: 10, Channels: 4, ByteWidth: 40, ColorMode: shared.ColorModeRGB, Data: []byte{0}}})
		registry.RegisterColorFinder(&fakeColorFinder{findSequence: []colorFindResult{{err: errors.New("missing")}, {match: shared.MatchResult[shared.Point]{Location: shared.Point{X: 2, Y: 3}, Confidence: 1}}}})
		scr := New(registry)

		result, err := scr.WaitFor(context.Background(), shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 4}), 50*time.Millisecond, time.Millisecond, &FindOptions{
			SearchRegion: &shared.Region{Left: 5, Top: 6, Width: 10, Height: 10},
		})
		if err != nil {
			t.Fatalf("unexpected wait success error: %v", err)
		}
		if point := result.(shared.Point); point != (shared.Point{X: 7, Y: 9}) {
			t.Fatalf("unexpected wait success point: %#v", point)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		registry := provider.NewRegistry()
		registry.RegisterScreen(&fakeScreenProvider{size: shared.Region{Left: 0, Top: 0, Width: 20, Height: 20}, regionImage: shared.Image{Width: 10, Height: 10, Channels: 4, ByteWidth: 40, ColorMode: shared.ColorModeRGB, Data: []byte{0}}})
		registry.RegisterColorFinder(&fakeColorFinder{matchErr: errors.New("still missing")})
		scr := New(registry)

		_, err := scr.WaitFor(context.Background(), shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 4}), 5*time.Millisecond, time.Millisecond, &FindOptions{
			SearchRegion: &shared.Region{Left: 0, Top: 0, Width: 10, Height: 10},
		})
		if err == nil {
			t.Fatal("expected wait timeout error")
		}
		if !strings.Contains(err.Error(), "timed out") || !strings.Contains(err.Error(), "still missing") {
			t.Fatalf("expected timeout error to include last error, got %v", err)
		}
	})
}

func TestScreenSmokeHighLevelAPIsEndToEnd(t *testing.T) {
	registry := provider.NewRegistry()
	screenProvider := &fakeScreenProvider{
		size: shared.Region{Left: 0, Top: 0, Width: 10, Height: 10},
		fullImage: shared.Image{
			Width:        20,
			Height:       20,
			Channels:     4,
			ByteWidth:    80,
			ColorMode:    shared.ColorModeRGB,
			PixelDensity: shared.PixelDensity{ScaleX: 2, ScaleY: 2},
			Data:         make([]byte, 20*20*4),
		},
	}
	windowProvider := &fakeWindowProvider{
		windows: []shared.WindowHandle{11},
		active:  11,
		titles:  map[shared.WindowHandle]string{11: "Editor"},
		regions: map[shared.WindowHandle]shared.Region{11: {Left: -2, Top: 1, Width: 8, Height: 12}},
	}
	imageProcessor := &fakeImageProcessor{color: shared.RGBA{R: 9, G: 8, B: 7, A: 255}}
	imageWriter := &fakeImageWriter{}
	registry.RegisterScreen(screenProvider)
	registry.RegisterWindow(windowProvider)
	registry.RegisterWindowFinder(&fakeWindowFinder{handles: []shared.WindowHandle{11}})
	registry.RegisterImageProcessor(imageProcessor)
	registry.RegisterImageWriter(imageWriter)
	scr := New(registry)

	width, err := scr.Width(context.Background())
	if err != nil || width != 10 {
		t.Fatalf("unexpected width result: width=%d err=%v", width, err)
	}
	height, err := scr.Height(context.Background())
	if err != nil || height != 10 {
		t.Fatalf("unexpected height result: height=%d err=%v", height, err)
	}

	found, err := scr.Find(context.Background(), shared.NewWindowQuery("editor", shared.MatchString("Editor")), nil)
	if err != nil {
		t.Fatalf("unexpected screen find window error: %v", err)
	}
	wrapped, ok := found.(*window.Window)
	if !ok {
		t.Fatalf("unexpected screen window result type: %T", found)
	}
	title, err := wrapped.Title(context.Background())
	if err != nil || title != "Editor" {
		t.Fatalf("unexpected wrapped title: title=%q err=%v", title, err)
	}
	region, err := wrapped.Region(context.Background())
	if err != nil {
		t.Fatalf("unexpected wrapped region error: %v", err)
	}
	if region != (shared.Region{Left: 0, Top: 1, Width: 6, Height: 9}) {
		t.Fatalf("unexpected wrapped region: %#v", region)
	}

	active, err := window.GetActiveWindow(context.Background(), registry)
	if err != nil || active.Handle != 11 {
		t.Fatalf("unexpected active window: %#v err=%v", active, err)
	}

	color, err := scr.ColorAt(context.Background(), shared.Point{X: 2, Y: 3})
	if err != nil {
		t.Fatalf("unexpected color at error: %v", err)
	}
	if color != (shared.RGBA{R: 9, G: 8, B: 7, A: 255}) {
		t.Fatalf("unexpected color at result: %#v", color)
	}
	if imageProcessor.lastPoint != (shared.Point{X: 4, Y: 6}) {
		t.Fatalf("unexpected physical color lookup point: %#v", imageProcessor.lastPoint)
	}

	path, err := scr.Capture(context.Background(), "capture.png")
	if err != nil {
		t.Fatalf("unexpected capture error: %v", err)
	}
	if path != "capture.png" || len(imageWriter.stored) != 1 || imageWriter.stored[0].Path != "capture.png" {
		t.Fatalf("unexpected capture output: path=%q stored=%#v", path, imageWriter.stored)
	}
}
