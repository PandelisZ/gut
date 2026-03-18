package screen

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
	"github.com/PandelisZ/gut/util"
	"github.com/PandelisZ/gut/window"
)

var errMatchNotFound = errors.New("screen search match not found")

type Config struct {
	Confidence        float64
	AutoHighlight     bool
	HighlightDuration time.Duration
	HighlightOpacity  float64
	ResourceDirectory string
}

type FindOptions struct {
	SearchRegion *shared.Region
	Confidence   *float64
	ProviderData any
}

type Screen struct {
	registry *provider.Registry
	Config   Config
}

func New(registry *provider.Registry) *Screen {
	if registry == nil {
		registry = provider.NewRegistry()
	}

	return &Screen{
		registry: registry,
		Config: Config{
			Confidence:        0.99,
			HighlightDuration: 500 * time.Millisecond,
			HighlightOpacity:  0.5,
		},
	}
}

func (s *Screen) Width(ctx context.Context) (int, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return 0, err
	}
	return provider.ScreenWidth(ctx)
}

func (s *Screen) Height(ctx context.Context) (int, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return 0, err
	}
	return provider.ScreenHeight(ctx)
}

func (s *Screen) Find(ctx context.Context, input any, opts *FindOptions) (any, error) {
	switch needle := input.(type) {
	case shared.Image:
		return s.findImage(ctx, needle, opts)
	case *shared.Image:
		if needle == nil {
			return nil, fmt.Errorf("image input is nil")
		}
		return s.findImage(ctx, *needle, opts)
	case shared.TextQuery:
		return s.findText(ctx, needle, opts)
	case *shared.TextQuery:
		if needle == nil {
			return nil, fmt.Errorf("text query input is nil")
		}
		return s.findText(ctx, *needle, opts)
	case shared.ColorQuery:
		return s.findColor(ctx, needle, opts)
	case *shared.ColorQuery:
		if needle == nil {
			return nil, fmt.Errorf("color query input is nil")
		}
		return s.findColor(ctx, *needle, opts)
	case shared.WindowQuery:
		return s.findWindow(ctx, needle, opts)
	case *shared.WindowQuery:
		if needle == nil {
			return nil, fmt.Errorf("window query input is nil")
		}
		return s.findWindow(ctx, *needle, opts)
	default:
		return nil, fmt.Errorf("unsupported screen find input type %T", input)
	}
}

func (s *Screen) FindAll(ctx context.Context, input any, opts *FindOptions) ([]any, error) {
	switch needle := input.(type) {
	case shared.Image:
		return s.findAllImages(ctx, needle, opts)
	case *shared.Image:
		if needle == nil {
			return nil, fmt.Errorf("image input is nil")
		}
		return s.findAllImages(ctx, *needle, opts)
	case shared.TextQuery:
		return s.findAllText(ctx, needle, opts)
	case *shared.TextQuery:
		if needle == nil {
			return nil, fmt.Errorf("text query input is nil")
		}
		return s.findAllText(ctx, *needle, opts)
	case shared.ColorQuery:
		return s.findAllColors(ctx, needle, opts)
	case *shared.ColorQuery:
		if needle == nil {
			return nil, fmt.Errorf("color query input is nil")
		}
		return s.findAllColors(ctx, *needle, opts)
	case shared.WindowQuery:
		return s.findAllWindows(ctx, needle, opts)
	case *shared.WindowQuery:
		if needle == nil {
			return nil, fmt.Errorf("window query input is nil")
		}
		return s.findAllWindows(ctx, *needle, opts)
	default:
		return nil, fmt.Errorf("unsupported screen find input type %T", input)
	}
}

func (s *Screen) WaitFor(ctx context.Context, input any, timeout time.Duration, interval time.Duration, opts *FindOptions) (any, error) {
	if timeout <= 0 {
		return s.Find(ctx, input, opts)
	}

	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		result, err := s.Find(ctx, input, opts)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for screen match after %s: %w", timeout, lastErr)
		}
		if err := util.SleepContext(ctx, interval); err != nil {
			return nil, err
		}
	}
}

func (s *Screen) Highlight(ctx context.Context, region shared.Region) (shared.Region, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return shared.Region{}, err
	}
	if err := provider.HighlightScreenRegion(ctx, region, s.Config.HighlightDuration, s.Config.HighlightOpacity); err != nil {
		return shared.Region{}, err
	}
	return region, nil
}

func (s *Screen) Grab(ctx context.Context) (shared.Image, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return shared.Image{}, err
	}
	return provider.GrabScreen(ctx)
}

func (s *Screen) GrabRegion(ctx context.Context, region shared.Region) (shared.Image, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return shared.Image{}, err
	}
	return provider.GrabScreenRegion(ctx, region)
}

func (s *Screen) Capture(ctx context.Context, path string) (string, error) {
	image, err := s.Grab(ctx)
	if err != nil {
		return "", err
	}
	return s.writeImage(ctx, path, image)
}

func (s *Screen) CaptureRegion(ctx context.Context, path string, region shared.Region) (string, error) {
	image, err := s.GrabRegion(ctx, region)
	if err != nil {
		return "", err
	}
	return s.writeImage(ctx, path, image)
}

func (s *Screen) ColorAt(ctx context.Context, point shared.Point) (shared.RGBA, error) {
	image, err := s.Grab(ctx)
	if err != nil {
		return shared.RGBA{}, err
	}
	processor, err := s.registry.ImageProcessor()
	if err != nil {
		return shared.RGBA{}, err
	}
	density := image.NormalizedPixelDensity()
	physical := shared.Point{
		X: int(math.Round(float64(point.X) * density.ScaleX)),
		Y: int(math.Round(float64(point.Y) * density.ScaleY)),
	}
	return processor.ColorAt(ctx, image, physical)
}

func (s *Screen) findImage(ctx context.Context, needle shared.Image, opts *FindOptions) (shared.Region, error) {
	finder, err := s.registry.ImageFinder()
	if err != nil {
		return shared.Region{}, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return shared.Region{}, err
	}
	match, err := finder.FindMatch(ctx, shared.MatchRequest[shared.Image, any]{
		Haystack:     haystack,
		Needle:       needle,
		Confidence:   s.confidence(opts, s.Config.Confidence),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return shared.Region{}, err
	}
	region := offsetRegion(match.Location, searchRegion)
	if err := s.highlightIfNeeded(ctx, region); err != nil {
		return shared.Region{}, err
	}
	return region, nil
}

func (s *Screen) findAllImages(ctx context.Context, needle shared.Image, opts *FindOptions) ([]any, error) {
	finder, err := s.registry.ImageFinder()
	if err != nil {
		return nil, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return nil, err
	}
	matches, err := finder.FindMatches(ctx, shared.MatchRequest[shared.Image, any]{
		Haystack:     haystack,
		Needle:       needle,
		Confidence:   s.confidence(opts, s.Config.Confidence),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return nil, err
	}
	results := make([]any, 0, len(matches))
	for _, match := range matches {
		region := offsetRegion(match.Location, searchRegion)
		if err := s.highlightIfNeeded(ctx, region); err != nil {
			return nil, err
		}
		results = append(results, region)
	}
	return results, nil
}

func (s *Screen) findText(ctx context.Context, query shared.TextQuery, opts *FindOptions) (shared.Region, error) {
	finder, err := s.registry.TextFinder()
	if err != nil {
		return shared.Region{}, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return shared.Region{}, err
	}
	match, err := finder.FindMatch(ctx, shared.MatchRequest[shared.TextQuery, any]{
		Haystack:     haystack,
		Needle:       query,
		Confidence:   s.confidence(opts, s.Config.Confidence),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return shared.Region{}, err
	}
	region := offsetRegion(match.Location, searchRegion)
	if err := s.highlightIfNeeded(ctx, region); err != nil {
		return shared.Region{}, err
	}
	return region, nil
}

func (s *Screen) findAllText(ctx context.Context, query shared.TextQuery, opts *FindOptions) ([]any, error) {
	finder, err := s.registry.TextFinder()
	if err != nil {
		return nil, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return nil, err
	}
	matches, err := finder.FindMatches(ctx, shared.MatchRequest[shared.TextQuery, any]{
		Haystack:     haystack,
		Needle:       query,
		Confidence:   s.confidence(opts, s.Config.Confidence),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return nil, err
	}
	results := make([]any, 0, len(matches))
	for _, match := range matches {
		region := offsetRegion(match.Location, searchRegion)
		if err := s.highlightIfNeeded(ctx, region); err != nil {
			return nil, err
		}
		results = append(results, region)
	}
	return results, nil
}

func (s *Screen) findColor(ctx context.Context, query shared.ColorQuery, opts *FindOptions) (shared.Point, error) {
	finder, err := s.registry.ColorFinder()
	if err != nil {
		return shared.Point{}, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return shared.Point{}, err
	}
	match, err := finder.FindMatch(ctx, shared.MatchRequest[shared.ColorQuery, any]{
		Haystack:     haystack,
		Needle:       query,
		Confidence:   s.confidence(opts, 1),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return shared.Point{}, err
	}
	return offsetPoint(match.Location, searchRegion), nil
}

func (s *Screen) findAllColors(ctx context.Context, query shared.ColorQuery, opts *FindOptions) ([]any, error) {
	finder, err := s.registry.ColorFinder()
	if err != nil {
		return nil, err
	}
	searchRegion, haystack, err := s.searchRegionAndHaystack(ctx, opts)
	if err != nil {
		return nil, err
	}
	matches, err := finder.FindMatches(ctx, shared.MatchRequest[shared.ColorQuery, any]{
		Haystack:     haystack,
		Needle:       query,
		Confidence:   s.confidence(opts, 1),
		ProviderData: providerDataPtr(opts),
	})
	if err != nil {
		return nil, err
	}
	results := make([]any, 0, len(matches))
	for _, match := range matches {
		results = append(results, offsetPoint(match.Location, searchRegion))
	}
	return results, nil
}

func (s *Screen) findWindow(ctx context.Context, query shared.WindowQuery, opts *FindOptions) (*window.Window, error) {
	matches, err := s.findAllWindows(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errMatchNotFound
	}
	result, ok := matches[0].(*window.Window)
	if !ok {
		return nil, fmt.Errorf("unexpected window search result type %T", matches[0])
	}
	return result, nil
}

func (s *Screen) findAllWindows(ctx context.Context, query shared.WindowQuery, opts *FindOptions) ([]any, error) {
	finder, err := s.registry.WindowFinder()
	if err != nil {
		return nil, err
	}
	searchRegion, err := s.resolveSearchRegion(ctx, opts)
	if err != nil {
		return nil, err
	}
	handles, err := finder.FindMatches(ctx, query)
	if err != nil {
		return nil, err
	}
	results := make([]any, 0, len(handles))
	for _, handle := range handles {
		candidate := window.New(s.registry, handle)
		region, err := candidate.Region(ctx)
		if err != nil {
			return nil, err
		}
		if !regionsOverlap(region, searchRegion) {
			continue
		}
		results = append(results, candidate)
	}
	return results, nil
}

func (s *Screen) searchRegionAndHaystack(ctx context.Context, opts *FindOptions) (shared.Region, shared.Image, error) {
	searchRegion, err := s.resolveSearchRegion(ctx, opts)
	if err != nil {
		return shared.Region{}, shared.Image{}, err
	}
	haystack, err := s.GrabRegion(ctx, searchRegion)
	if err != nil {
		return shared.Region{}, shared.Image{}, err
	}
	return searchRegion, haystack, nil
}

func (s *Screen) resolveSearchRegion(ctx context.Context, opts *FindOptions) (shared.Region, error) {
	provider, err := s.registry.Screen()
	if err != nil {
		return shared.Region{}, err
	}
	bounds, err := provider.ScreenSize(ctx)
	if err != nil {
		return shared.Region{}, err
	}
	if opts == nil || opts.SearchRegion == nil {
		return bounds, nil
	}
	region := *opts.SearchRegion
	if err := validateSearchRegion(region, bounds); err != nil {
		return shared.Region{}, err
	}
	return region, nil
}

func (s *Screen) confidence(opts *FindOptions, fallback float64) *float64 {
	if opts != nil && opts.Confidence != nil {
		return opts.Confidence
	}
	value := fallback
	return &value
}

func (s *Screen) writeImage(ctx context.Context, path string, image shared.Image) (string, error) {
	writer, err := s.registry.ImageWriter()
	if err != nil {
		return "", err
	}
	if err := writer.Store(ctx, provider.ImageWriterParameters{Image: image, Path: path}); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Screen) highlightIfNeeded(ctx context.Context, region shared.Region) error {
	if !s.Config.AutoHighlight {
		return nil
	}
	_, err := s.Highlight(ctx, region)
	return err
}

func validateSearchRegion(region shared.Region, bounds shared.Region) error {
	if region.Left < 0 || region.Top < 0 {
		return fmt.Errorf("invalid search region %s: origin must be non-negative", region)
	}
	if region.Width < 0 || region.Height < 0 {
		return fmt.Errorf("invalid search region %s: dimensions must be non-negative", region)
	}
	if region.Width < 2 || region.Height < 2 {
		return fmt.Errorf("invalid search region %s: width and height must be at least 2", region)
	}
	if region.Left+region.Width > bounds.Left+bounds.Width || region.Top+region.Height > bounds.Top+bounds.Height {
		return fmt.Errorf("invalid search region %s: region must remain within screen bounds %s", region, bounds)
	}
	return nil
}

func offsetRegion(region shared.Region, searchRegion shared.Region) shared.Region {
	return shared.Region{
		Left:   region.Left + searchRegion.Left,
		Top:    region.Top + searchRegion.Top,
		Width:  region.Width,
		Height: region.Height,
	}
}

func offsetPoint(point shared.Point, searchRegion shared.Region) shared.Point {
	return shared.Point{X: point.X + searchRegion.Left, Y: point.Y + searchRegion.Top}
}

func regionsOverlap(a, b shared.Region) bool {
	return a.Left < b.Left+b.Width && b.Left < a.Left+a.Width && a.Top < b.Top+b.Height && b.Top < a.Top+a.Height
}

func providerDataPtr(opts *FindOptions) *any {
	if opts == nil || opts.ProviderData == nil {
		return nil
	}
	value := opts.ProviderData
	return &value
}
