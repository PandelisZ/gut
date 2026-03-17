package libnutcore

import (
	"time"

	"gut/native/common"
)

const BackendName = "libnut-core"

type BindingState string

const (
	BindingStateLinked      BindingState = "linked"
	BindingStateUnavailable BindingState = "unavailable"
)

type BackendInfo struct {
	Name         string
	Platform     string
	BindingState BindingState
	Notes        []string
}

type Options struct {
	MouseDelay    time.Duration
	KeyboardDelay time.Duration
	XDisplayName  string
}

func DefaultOptions() Options {
	return Options{
		MouseDelay:    10 * time.Millisecond,
		KeyboardDelay: 10 * time.Millisecond,
	}
}

type Client interface {
	Info() BackendInfo
	Capabilities() common.CapabilitySet
	GetPermissionSnapshot() (common.PermissionSnapshot, error)
	GetFocusedWindow() (common.FocusedWindowMetadata, error)
	RaiseFocusedWindow() error
	GetFocusedElement() (common.UIElementMetadata, error)
	PerformFocusedElementAction(action common.AXAction) error
	GetElementAtPoint(position common.Point) (common.UIElementMetadata, error)
	PerformElementActionAtPoint(position common.Point, action common.AXAction) error
	FocusElementAtPoint(position common.Point) error

	DragMouse(position common.Point, button common.MouseButton) error
	MoveMouse(position common.Point) error
	GetMousePosition() (common.Point, error)
	MouseClick(button common.MouseButton, double bool) error
	MouseToggle(state common.ButtonState, button common.MouseButton) error
	ScrollMouse(horizontal, vertical int) error
	SetMouseDelay(delay time.Duration) error

	KeyTap(key string, modifiers ...string) error
	KeyToggle(key string, state common.KeyState, modifiers ...string) error
	TypeString(text string) error
	TypeStringDelayed(text string, cpm int) error
	SetKeyboardDelay(delay time.Duration) error

	GetScreenSize() (common.Size, error)
	Highlight(region common.Rect, duration time.Duration, opacity float64) error
	CaptureScreen(region *common.Rect) (*common.Bitmap, error)

	GetWindows() ([]common.WindowHandle, error)
	GetActiveWindow() (common.WindowHandle, error)
	GetWindowRect(handle common.WindowHandle) (common.Rect, error)
	GetWindowTitle(handle common.WindowHandle) (string, error)
	FocusWindow(handle common.WindowHandle) (bool, error)
	ResizeWindow(handle common.WindowHandle, size common.Size) (bool, error)
	MoveWindow(handle common.WindowHandle, origin common.Point) (bool, error)
	MinimizeWindow(handle common.WindowHandle) (bool, error)
	RestoreWindow(handle common.WindowHandle) (bool, error)

	GetXDisplayName() (string, error)
	SetXDisplayName(name string) error
}

func New(options Options) Client {
	return newClient(withDefaults(options))
}

func withDefaults(options Options) Options {
	defaults := DefaultOptions()
	if options.MouseDelay == 0 {
		options.MouseDelay = defaults.MouseDelay
	}
	if options.KeyboardDelay == 0 {
		options.KeyboardDelay = defaults.KeyboardDelay
	}
	return options
}
