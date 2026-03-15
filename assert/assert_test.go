package assert

import (
	"context"
	"errors"
	"testing"
	"time"

	"gut/provider"
	"gut/screen"
	"gut/shared"
)

var assertTestBinarySalt = "gut-assert-test-salt-1"

type fakeScreenProvider struct {
	size        shared.Region
	regionImage shared.Image
}

func (f *fakeScreenProvider) GrabScreen(context.Context) (shared.Image, error) {
	return f.regionImage, nil
}

func (f *fakeScreenProvider) GrabScreenRegion(context.Context, shared.Region) (shared.Image, error) {
	return f.regionImage, nil
}

func (f *fakeScreenProvider) HighlightScreenRegion(context.Context, shared.Region, time.Duration, float64) error {
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
	allMatches []shared.MatchResult[shared.Point]
	matchErr   error
}

func (f *fakeColorFinder) FindMatch(context.Context, shared.MatchRequest[shared.ColorQuery, any]) (shared.MatchResult[shared.Point], error) {
	if f.matchErr != nil {
		return shared.MatchResult[shared.Point]{Error: f.matchErr}, f.matchErr
	}
	if len(f.allMatches) == 0 {
		return shared.MatchResult[shared.Point]{Error: errors.New("not found")}, errors.New("not found")
	}
	return f.allMatches[0], nil
}

func (f *fakeColorFinder) FindMatches(context.Context, shared.MatchRequest[shared.ColorQuery, any]) ([]shared.MatchResult[shared.Point], error) {
	if f.matchErr != nil {
		return nil, f.matchErr
	}
	return append([]shared.MatchResult[shared.Point](nil), f.allMatches...), nil
}

func TestAssertVisibleAndNotVisible(t *testing.T) {
	newAssert := func(matches []shared.MatchResult[shared.Point], matchErr error) *Assert {
		registry := provider.NewRegistry()
		registry.RegisterScreen(&fakeScreenProvider{
			size:        shared.Region{Left: 0, Top: 0, Width: 20, Height: 20},
			regionImage: shared.Image{Width: 20, Height: 20, Channels: 4, ByteWidth: 80, ColorMode: shared.ColorModeRGB, Data: []byte{0}},
		})
		registry.RegisterColorFinder(&fakeColorFinder{allMatches: matches, matchErr: matchErr})
		return New(screen.New(registry))
	}

	query := shared.NewColorQuery("needle", shared.RGBA{R: 1, G: 2, B: 3, A: 255})
	searchRegion := &shared.Region{Left: 0, Top: 0, Width: 10, Height: 10}

	if err := newAssert([]shared.MatchResult[shared.Point]{{Location: shared.Point{X: 1, Y: 2}, Confidence: 1}}, nil).Visible(context.Background(), query, searchRegion, nil); err != nil {
		t.Fatalf("expected visible assertion to succeed, got %v", err)
	}
	if err := newAssert(nil, errors.New("not found")).Visible(context.Background(), query, searchRegion, nil); err == nil {
		t.Fatal("expected visible assertion to fail when no match exists")
	}
	if err := newAssert(nil, nil).NotVisible(context.Background(), query, searchRegion, nil); err != nil {
		t.Fatalf("expected not visible assertion to succeed, got %v", err)
	}
	if err := newAssert([]shared.MatchResult[shared.Point]{{Location: shared.Point{X: 1, Y: 2}, Confidence: 1}}, nil).NotVisible(context.Background(), query, searchRegion, nil); err == nil {
		t.Fatal("expected not visible assertion to fail when a match exists")
	}
}
