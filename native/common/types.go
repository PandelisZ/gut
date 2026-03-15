package common

type Point struct {
	X int
	Y int
}

type Size struct {
	Width  int
	Height int
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Bitmap struct {
	Width         int
	Height        int
	ByteWidth     int
	BitsPerPixel  int
	BytesPerPixel int
	Image         []byte
}

type WindowHandle int64

type Window struct {
	Handle WindowHandle
	Title  string
	Rect   Rect
}

type MouseButton string

const (
	MouseButtonLeft   MouseButton = "left"
	MouseButtonRight  MouseButton = "right"
	MouseButtonMiddle MouseButton = "middle"
)

type ButtonState string

const (
	ButtonStateUp   ButtonState = "up"
	ButtonStateDown ButtonState = "down"
)

type KeyState string

const (
	KeyStateUp   KeyState = "up"
	KeyStateDown KeyState = "down"
)

type KeyModifier string

const (
	KeyModifierAlt     KeyModifier = "alt"
	KeyModifierControl KeyModifier = "control"
	KeyModifierShift   KeyModifier = "shift"
	KeyModifierMeta    KeyModifier = "meta"
	KeyModifierFn      KeyModifier = "fn"
	KeyModifierNone    KeyModifier = "none"
)
