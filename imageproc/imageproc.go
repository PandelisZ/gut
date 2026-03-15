package imageproc

import (
	"context"
	"errors"
	"fmt"
	"math"

	"gut/provider"
	"gut/shared"
)

var ErrPointOutOfBounds = errors.New("image point out of bounds")
var ErrUnsupportedChannelLayout = errors.New("unsupported image channel layout")
var ErrColorMatchNotFound = errors.New("color match not found")
var ErrWindowMatchNotFound = errors.New("window match not found")
var ErrImageFinderUnavailable = errors.New("image finder unavailable")
var ErrTextFinderUnavailable = errors.New("text finder unavailable")
var ErrWindowFinderUnavailable = errors.New("window finder unavailable")

type Processor struct{}

type ColorFinder struct {
	processor *Processor
}

type WindowFinder struct {
	windows provider.WindowProvider
}

type UnavailableImageFinder struct{}

type UnavailableTextFinder struct{}

type UnavailableWindowFinder struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func NewColorFinder(processor *Processor) *ColorFinder {
	if processor == nil {
		processor = NewProcessor()
	}
	return &ColorFinder{processor: processor}
}

func NewWindowFinder(windows provider.WindowProvider) *WindowFinder {
	return &WindowFinder{windows: windows}
}

func NewUnavailableImageFinder() *UnavailableImageFinder {
	return &UnavailableImageFinder{}
}

func NewUnavailableTextFinder() *UnavailableTextFinder {
	return &UnavailableTextFinder{}
}

func NewUnavailableWindowFinder() *UnavailableWindowFinder {
	return &UnavailableWindowFinder{}
}

func (p *Processor) ColorAt(ctx context.Context, image shared.Image, location shared.Point) (shared.RGBA, error) {
	if err := ctx.Err(); err != nil {
		return shared.RGBA{}, err
	}
	return colorAt(image, location)
}

func (f *ColorFinder) FindMatch(ctx context.Context, request shared.MatchRequest[shared.ColorQuery, any]) (shared.MatchResult[shared.Point], error) {
	matches, err := f.FindMatches(ctx, request)
	if err != nil {
		return shared.MatchResult[shared.Point]{Error: err}, err
	}
	if len(matches) == 0 {
		return shared.MatchResult[shared.Point]{Error: ErrColorMatchNotFound}, ErrColorMatchNotFound
	}
	return matches[0], nil
}

func (f *ColorFinder) FindMatches(ctx context.Context, request shared.MatchRequest[shared.ColorQuery, any]) ([]shared.MatchResult[shared.Point], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if request.Confidence != nil && *request.Confidence > 1 {
		return nil, nil
	}

	var matches []shared.MatchResult[shared.Point]
	for y := 0; y < request.Haystack.Height; y++ {
		for x := 0; x < request.Haystack.Width; x++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			pixel, err := colorAt(request.Haystack, shared.Point{X: x, Y: y})
			if err != nil {
				return nil, err
			}
			if pixel != request.Needle.By.Color {
				continue
			}
			matches = append(matches, shared.MatchResult[shared.Point]{
				Confidence: 1,
				Location:   logicalPoint(request.Haystack.PixelDensity, shared.Point{X: x, Y: y}),
			})
		}
	}
	return matches, nil
}

func (f *WindowFinder) FindMatch(ctx context.Context, query shared.WindowQuery) (shared.WindowHandle, error) {
	matches, err := f.FindMatches(ctx, query)
	if err != nil {
		return 0, err
	}
	if len(matches) == 0 {
		return 0, ErrWindowMatchNotFound
	}
	return matches[0], nil
}

func (f *WindowFinder) FindMatches(ctx context.Context, query shared.WindowQuery) ([]shared.WindowHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.windows == nil {
		return nil, ErrWindowFinderUnavailable
	}

	windows, err := f.windows.GetWindows(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]shared.WindowHandle, 0, len(windows))
	for _, window := range windows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		title, err := f.windows.GetWindowTitle(ctx, window)
		if err != nil {
			return nil, err
		}
		if query.By.Title.Match(title) {
			matches = append(matches, window)
		}
	}
	return matches, nil
}

func (f *UnavailableImageFinder) FindMatch(context.Context, shared.MatchRequest[shared.Image, any]) (shared.MatchResult[shared.Region], error) {
	return shared.MatchResult[shared.Region]{Error: ErrImageFinderUnavailable}, ErrImageFinderUnavailable
}

func (f *UnavailableImageFinder) FindMatches(context.Context, shared.MatchRequest[shared.Image, any]) ([]shared.MatchResult[shared.Region], error) {
	return nil, ErrImageFinderUnavailable
}

func (f *UnavailableTextFinder) FindMatch(context.Context, shared.MatchRequest[shared.TextQuery, any]) (shared.MatchResult[shared.Region], error) {
	return shared.MatchResult[shared.Region]{Error: ErrTextFinderUnavailable}, ErrTextFinderUnavailable
}

func (f *UnavailableTextFinder) FindMatches(context.Context, shared.MatchRequest[shared.TextQuery, any]) ([]shared.MatchResult[shared.Region], error) {
	return nil, ErrTextFinderUnavailable
}

func (f *UnavailableWindowFinder) FindMatch(context.Context, shared.WindowQuery) (shared.WindowHandle, error) {
	return 0, ErrWindowFinderUnavailable
}

func (f *UnavailableWindowFinder) FindMatches(context.Context, shared.WindowQuery) ([]shared.WindowHandle, error) {
	return nil, ErrWindowFinderUnavailable
}

func colorAt(image shared.Image, location shared.Point) (shared.RGBA, error) {
	if location.X < 0 || location.Y < 0 || location.X >= image.Width || location.Y >= image.Height {
		return shared.RGBA{}, ErrPointOutOfBounds
	}

	stride := image.ByteWidth
	if stride == 0 {
		stride = image.Width * image.Channels
	}
	offset := location.Y*stride + location.X*image.Channels
	if offset < 0 || offset+image.Channels > len(image.Data) {
		return shared.RGBA{}, fmt.Errorf("image data too short for point %s", location)
	}

	color := shared.RGBA{A: 255}
	if image.Channels == 4 {
		color.A = image.Data[offset+3]
	} else if image.Channels != 3 {
		return shared.RGBA{}, fmt.Errorf("%w: %d", ErrUnsupportedChannelLayout, image.Channels)
	}

	switch image.ColorMode {
	case shared.ColorModeRGB:
		color.R = image.Data[offset]
		color.G = image.Data[offset+1]
		color.B = image.Data[offset+2]
	case shared.ColorModeBGR:
		color.B = image.Data[offset]
		color.G = image.Data[offset+1]
		color.R = image.Data[offset+2]
	default:
		return shared.RGBA{}, fmt.Errorf("unsupported color mode: %s", image.ColorMode)
	}
	return color, nil
}

func logicalPoint(density shared.PixelDensity, physical shared.Point) shared.Point {
	scaleX := density.ScaleX
	if scaleX == 0 {
		scaleX = 1
	}
	scaleY := density.ScaleY
	if scaleY == 0 {
		scaleY = 1
	}
	return shared.Point{
		X: int(math.Round(float64(physical.X) / scaleX)),
		Y: int(math.Round(float64(physical.Y) / scaleY)),
	}
}

var _ provider.ImageProcessor = (*Processor)(nil)
var _ provider.ColorFinder = (*ColorFinder)(nil)
var _ provider.WindowFinder = (*WindowFinder)(nil)
var _ provider.ImageFinder = (*UnavailableImageFinder)(nil)
var _ provider.TextFinder = (*UnavailableTextFinder)(nil)
var _ provider.WindowFinder = (*UnavailableWindowFinder)(nil)
