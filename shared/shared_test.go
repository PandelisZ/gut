package shared

import (
	"regexp"
	"testing"
)

func TestGeometryHelpers(t *testing.T) {
	point := Point{X: 10, Y: 20}
	if point.String() != "(10, 20)" {
		t.Fatalf("unexpected point string: %s", point.String())
	}
	if !IsPoint(point) {
		t.Fatal("expected point type guard to accept Point")
	}

	size := Size{Width: 3, Height: 7}
	if size.Area() != 21 {
		t.Fatalf("unexpected size area: %d", size.Area())
	}
	if size.String() != "(3x7)" {
		t.Fatalf("unexpected size string: %s", size.String())
	}

	region := Region{Left: 1, Top: 2, Width: 4, Height: 5}
	if region.Area() != 20 {
		t.Fatalf("unexpected region area: %d", region.Area())
	}
	if region.Origin() != (Point{X: 1, Y: 2}) {
		t.Fatalf("unexpected region origin: %+v", region.Origin())
	}
	if region.Size() != (Size{Width: 4, Height: 5}) {
		t.Fatalf("unexpected region size: %+v", region.Size())
	}
}

func TestEnumAndColorHelpers(t *testing.T) {
	if got := KeyAudioRandom.String(); got != "AudioRandom" {
		t.Fatalf("unexpected key string: %s", got)
	}
	if !KeyEscape.Valid() || Key(-1).Valid() {
		t.Fatal("unexpected key validity result")
	}
	if ButtonMiddle.String() != "middle" {
		t.Fatalf("unexpected button string: %s", ButtonMiddle.String())
	}
	color := RGBA{R: 1, G: 2, B: 3, A: 4}
	if color.Hex() != "#01020304" {
		t.Fatalf("unexpected rgba hex: %s", color.Hex())
	}
}

func TestImageValidationAndHelpers(t *testing.T) {
	image, err := NewImage(10, 12, []byte{1, 2, 3, 4}, 4, "img-1", 32, 40, ColorModeBGR, PixelDensity{})
	if err != nil {
		t.Fatalf("unexpected image error: %v", err)
	}
	if !image.HasAlphaChannel() {
		t.Fatal("expected alpha channel to be detected")
	}
	if density := image.NormalizedPixelDensity(); density.ScaleX != 1 || density.ScaleY != 1 {
		t.Fatalf("unexpected pixel density normalization: %+v", density)
	}

	if _, err := NewImage(1, 1, nil, 0, "bad", 32, 4, ColorModeRGB, PixelDensity{ScaleX: 1, ScaleY: 1}); err == nil {
		t.Fatal("expected invalid channel count to fail")
	}
}

func TestQueryHelpers(t *testing.T) {
	wordQuery := NewWordQuery("q1", "hello")
	if !wordQuery.Valid() || !wordQuery.IsWord() || IsLineQuery(wordQuery) {
		t.Fatal("word query helpers returned unexpected result")
	}

	lineQuery := NewLineQuery("q2", "hello world")
	if !lineQuery.Valid() || !IsLineQuery(lineQuery) || IsWordQuery(lineQuery) {
		t.Fatal("line query helpers returned unexpected result")
	}

	windowQuery := NewWindowQuery("q3", MatchPattern(regexp.MustCompile(`(?i)calculator`)))
	if !windowQuery.Valid() || !windowQuery.By.Title.Match("Calculator") {
		t.Fatal("window query matcher did not behave as expected")
	}

	description := WindowElementDescription{Role: "button", Title: ptr(MatchString("OK"))}
	windowElementQuery := NewWindowElementQuery("q4", description)
	if !windowElementQuery.Valid() || !IsWindowElementQuery(windowElementQuery) {
		t.Fatal("window element query helpers returned unexpected result")
	}

	colorQuery := NewColorQuery("q5", RGBA{R: 255, G: 0, B: 0, A: 255})
	if !IsColorQuery(colorQuery) {
		t.Fatal("expected color query type guard to accept ColorQuery")
	}
}

func TestSearchAndMatchHelpers(t *testing.T) {
	confidence := 0.9
	region := &Region{Left: 0, Top: 0, Width: 10, Height: 10}
	options := SearchOptions[string]{SearchRegion: region, Confidence: &confidence, ProviderData: ptr("provider")}
	if !options.HasSearchRegion() || !options.HasConfidence() || !options.HasProviderData() {
		t.Fatal("search options helper did not detect optional values")
	}

	result := MatchResult[Region]{Confidence: 1, Location: Region{Width: 1, Height: 1}}
	if !result.Succeeded() {
		t.Fatal("expected successful match result")
	}
}

func ptr[T any](value T) *T {
	return &value
}
