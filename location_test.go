package gut

import (
	"testing"

	"github.com/PandelisZ/gut/shared"
)

func TestCenterOf(t *testing.T) {
	center, err := CenterOf(shared.Region{Left: 10, Top: 20, Width: 5, Height: 7})
	if err != nil {
		t.Fatalf("unexpected CenterOf error: %v", err)
	}

	expected := shared.Point{X: 12, Y: 23}
	if center != expected {
		t.Fatalf("unexpected center: got %v want %v", center, expected)
	}
}

func TestCenterOfRejectsInvalidRegion(t *testing.T) {
	_, err := CenterOf(shared.Region{Left: 1, Top: 2, Width: -1, Height: 3})
	if err == nil {
		t.Fatal("expected invalid region error")
	}
}

func TestRandomPointInReturnsPointInsideRegion(t *testing.T) {
	region := shared.Region{Left: 10, Top: 20, Width: 5, Height: 7}
	for idx := 0; idx < 50; idx++ {
		point, err := RandomPointIn(region)
		if err != nil {
			t.Fatalf("unexpected RandomPointIn error: %v", err)
		}
		if point.X < region.Left || point.X >= region.Left+region.Width {
			t.Fatalf("point X out of bounds: %v", point)
		}
		if point.Y < region.Top || point.Y >= region.Top+region.Height {
			t.Fatalf("point Y out of bounds: %v", point)
		}
	}
}

func TestRandomPointInHandlesZeroSizedRegion(t *testing.T) {
	point, err := RandomPointIn(shared.Region{Left: 5, Top: 6, Width: 0, Height: 0})
	if err != nil {
		t.Fatalf("unexpected RandomPointIn error: %v", err)
	}
	if point != (shared.Point{X: 5, Y: 6}) {
		t.Fatalf("unexpected point: got %v", point)
	}
}
