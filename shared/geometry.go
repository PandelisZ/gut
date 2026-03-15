package shared

import "fmt"

type Point struct {
	X int
	Y int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func IsPoint(v any) bool {
	switch v.(type) {
	case Point, *Point:
		return true
	default:
		return false
	}
}

type Size struct {
	Width  int
	Height int
}

func (s Size) Area() int {
	return s.Width * s.Height
}

func (s Size) String() string {
	return fmt.Sprintf("(%dx%d)", s.Width, s.Height)
}

func (s Size) Valid() bool {
	return s.Width >= 0 && s.Height >= 0
}

func IsSize(v any) bool {
	switch v.(type) {
	case Size, *Size:
		return true
	default:
		return false
	}
}

type Region struct {
	Left   int
	Top    int
	Width  int
	Height int
}

func (r Region) Area() int {
	return r.Width * r.Height
}

func (r Region) String() string {
	return fmt.Sprintf("(%d, %d, %d, %d)", r.Left, r.Top, r.Width, r.Height)
}

func (r Region) Valid() bool {
	return r.Width >= 0 && r.Height >= 0
}

func (r Region) Origin() Point {
	return Point{X: r.Left, Y: r.Top}
}

func (r Region) Size() Size {
	return Size{Width: r.Width, Height: r.Height}
}

func IsRegion(v any) bool {
	switch v.(type) {
	case Region, *Region:
		return true
	default:
		return false
	}
}
