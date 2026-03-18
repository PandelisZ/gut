package gut

import (
	"fmt"
	"math/rand"

	"github.com/PandelisZ/gut/shared"
)

func CenterOf(region shared.Region) (shared.Point, error) {
	if !region.Valid() {
		return shared.Point{}, fmt.Errorf("invalid region: %v", region)
	}
	return shared.Point{
		X: region.Left + region.Width/2,
		Y: region.Top + region.Height/2,
	}, nil
}

func RandomPointIn(region shared.Region) (shared.Point, error) {
	if !region.Valid() {
		return shared.Point{}, fmt.Errorf("invalid region: %v", region)
	}

	x := region.Left
	if region.Width > 0 {
		x += rand.Intn(region.Width)
	}
	y := region.Top
	if region.Height > 0 {
		y += rand.Intn(region.Height)
	}

	return shared.Point{X: x, Y: y}, nil
}
