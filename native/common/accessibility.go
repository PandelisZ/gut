package common

type AXAction string

const (
	AXPress    AXAction = "AXPress"
	AXRaise    AXAction = "AXRaise"
	AXShowMenu AXAction = "AXShowMenu"
	AXConfirm  AXAction = "AXConfirm"
	AXPick     AXAction = "AXPick"
)

type PermissionStatus struct {
	Granted   bool
	Supported bool
	Reason    string
}

type PermissionSnapshot struct {
	Accessibility   PermissionStatus
	ScreenRecording PermissionStatus
}

type FocusedWindowMetadata struct {
	Handle    WindowHandle
	Title     string
	Role      string
	Subrole   string
	Rect      Rect
	RectKnown bool
	Focused   bool
	Main      bool
	Minimized bool
	OwnerPID  int
	OwnerName string
	BundleID  string
}

type UIElementMetadata struct {
	Role        string
	Subrole     string
	Title       string
	Description string
	Value       string
	Enabled     bool
	Focused     bool
	Frame       Rect
	FrameKnown  bool
	Actions     []string
}
