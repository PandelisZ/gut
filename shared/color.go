package shared

import "fmt"

type ColorMode int

const (
	ColorModeBGR ColorMode = iota
	ColorModeRGB
)

func (m ColorMode) String() string {
	switch m {
	case ColorModeBGR:
		return "BGR"
	case ColorModeRGB:
		return "RGB"
	default:
		return fmt.Sprintf("ColorMode(%d)", int(m))
	}
}

type RGBA struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (c RGBA) String() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
}

func (c RGBA) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
