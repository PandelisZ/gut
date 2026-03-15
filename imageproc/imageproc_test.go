package imageproc

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"gut/shared"
)

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

func TestProcessorColorAtSupportsRGBAndBGRLayouts(t *testing.T) {
	processor := NewProcessor()
	tests := []struct {
		name     string
		image    shared.Image
		location shared.Point
		want     shared.RGBA
	}{
		{
			name: "rgb with alpha",
			image: shared.Image{
				Width:     2,
				Height:    1,
				Channels:  4,
				ByteWidth: 8,
				ColorMode: shared.ColorModeRGB,
				Data:      []byte{1, 2, 3, 4, 10, 20, 30, 40},
			},
			location: shared.Point{X: 1, Y: 0},
			want:     shared.RGBA{R: 10, G: 20, B: 30, A: 40},
		},
		{
			name: "bgr without alpha",
			image: shared.Image{
				Width:     1,
				Height:    1,
				Channels:  3,
				ColorMode: shared.ColorModeBGR,
				Data:      []byte{30, 20, 10},
			},
			location: shared.Point{X: 0, Y: 0},
			want:     shared.RGBA{R: 10, G: 20, B: 30, A: 255},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := processor.ColorAt(context.Background(), test.image, test.location)
			if err != nil {
				t.Fatalf("color at failed: %v", err)
			}
			if got != test.want {
				t.Fatalf("unexpected color: got %#v want %#v", got, test.want)
			}
		})
	}
}

func TestProcessorColorAtReturnsStableErrors(t *testing.T) {
	processor := NewProcessor()
	image := shared.Image{
		Width:     1,
		Height:    1,
		Channels:  2,
		ColorMode: shared.ColorModeRGB,
		Data:      []byte{1, 2},
	}

	if _, err := processor.ColorAt(context.Background(), image, shared.Point{X: 1, Y: 0}); !errors.Is(err, ErrPointOutOfBounds) {
		t.Fatalf("expected bounds error, got %v", err)
	}
	if _, err := processor.ColorAt(context.Background(), image, shared.Point{X: 0, Y: 0}); !errors.Is(err, ErrUnsupportedChannelLayout) {
		t.Fatalf("expected channel layout error, got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	cancel()
	if _, err := processor.ColorAt(ctx, image, shared.Point{X: 0, Y: 0}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
}

func TestColorFinderFindsExactMatchesAndNormalizesDensity(t *testing.T) {
	finder := NewColorFinder(NewProcessor())
	image := shared.Image{
		Width:        4,
		Height:       3,
		Channels:     4,
		ByteWidth:    16,
		ColorMode:    shared.ColorModeBGR,
		PixelDensity: shared.PixelDensity{ScaleX: 2, ScaleY: 2},
		Data: []byte{
			0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
			0, 0, 0, 255, 7, 6, 5, 255, 0, 0, 0, 255, 7, 6, 5, 255,
			0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
		},
	}
	request := shared.MatchRequest[shared.ColorQuery, any]{
		Haystack: image,
		Needle:   shared.NewColorQuery("needle", shared.RGBA{R: 5, G: 6, B: 7, A: 255}),
	}

	matches, err := finder.FindMatches(context.Background(), request)
	if err != nil {
		t.Fatalf("find matches failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Confidence != 1 || matches[1].Confidence != 1 {
		t.Fatalf("expected confidence 1.0 for matches, got %#v", matches)
	}
	if matches[0].Location != (shared.Point{X: 1, Y: 1}) {
		t.Fatalf("unexpected first logical location: %#v", matches[0].Location)
	}
	if matches[1].Location != (shared.Point{X: 2, Y: 1}) {
		t.Fatalf("unexpected second logical location: %#v", matches[1].Location)
	}

	first, err := finder.FindMatch(context.Background(), request)
	if err != nil {
		t.Fatalf("find match failed: %v", err)
	}
	if first.Location != matches[0].Location {
		t.Fatalf("expected first match location %#v, got %#v", matches[0].Location, first.Location)
	}
}

func TestColorFinderReturnsStableNoMatchError(t *testing.T) {
	finder := NewColorFinder(nil)
	_, err := finder.FindMatch(context.Background(), shared.MatchRequest[shared.ColorQuery, any]{
		Haystack: shared.Image{
			Width:     1,
			Height:    1,
			Channels:  4,
			ByteWidth: 4,
			ColorMode: shared.ColorModeRGB,
			Data:      []byte{0, 0, 0, 255},
		},
		Needle: shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 255}),
	})
	if !errors.Is(err, ErrColorMatchNotFound) {
		t.Fatalf("expected no-match error, got %v", err)
	}
}

func TestWindowFinderFiltersTitlesWithStringMatcher(t *testing.T) {
	windows := &stubWindowProvider{
		windows: []shared.WindowHandle{1, 2, 3},
		titles: map[shared.WindowHandle]string{
			1: "Terminal",
			2: "Editor",
			3: "Terminal - Logs",
		},
	}
	finder := NewWindowFinder(windows)
	query := shared.NewWindowQuery("terminal", shared.MatchPattern(regexp.MustCompile(`^Terminal`)))

	matches, err := finder.FindMatches(context.Background(), query)
	if err != nil {
		t.Fatalf("find matches failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0] != 1 || matches[1] != 3 {
		t.Fatalf("unexpected matches: %#v", matches)
	}

	first, err := finder.FindMatch(context.Background(), query)
	if err != nil {
		t.Fatalf("find match failed: %v", err)
	}
	if first != 1 {
		t.Fatalf("expected first match handle 1, got %d", first)
	}
}

func TestWindowFinderReturnsStableNoMatchAndUnavailableErrors(t *testing.T) {
	finder := NewWindowFinder(&stubWindowProvider{windows: []shared.WindowHandle{1}, titles: map[shared.WindowHandle]string{1: "Editor"}})
	_, err := finder.FindMatch(context.Background(), shared.NewWindowQuery("terminal", shared.MatchString("Terminal")))
	if !errors.Is(err, ErrWindowMatchNotFound) {
		t.Fatalf("expected no-match error, got %v", err)
	}

	_, err = NewUnavailableWindowFinder().FindMatch(context.Background(), shared.NewWindowQuery("terminal", shared.MatchString("Terminal")))
	if !errors.Is(err, ErrWindowFinderUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestUnavailableFindersReturnStableErrors(t *testing.T) {
	imageFinder := NewUnavailableImageFinder()
	if _, err := imageFinder.FindMatch(context.Background(), shared.MatchRequest[shared.Image, any]{}); !errors.Is(err, ErrImageFinderUnavailable) {
		t.Fatalf("expected image finder unavailable error, got %v", err)
	}
	if _, err := imageFinder.FindMatches(context.Background(), shared.MatchRequest[shared.Image, any]{}); !errors.Is(err, ErrImageFinderUnavailable) {
		t.Fatalf("expected image finder unavailable error, got %v", err)
	}

	textFinder := NewUnavailableTextFinder()
	if _, err := textFinder.FindMatch(context.Background(), shared.MatchRequest[shared.TextQuery, any]{}); !errors.Is(err, ErrTextFinderUnavailable) {
		t.Fatalf("expected text finder unavailable error, got %v", err)
	}
	if _, err := textFinder.FindMatches(context.Background(), shared.MatchRequest[shared.TextQuery, any]{}); !errors.Is(err, ErrTextFinderUnavailable) {
		t.Fatalf("expected text finder unavailable error, got %v", err)
	}
}
