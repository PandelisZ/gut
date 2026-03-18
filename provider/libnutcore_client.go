package provider

import (
	"time"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/native/libnutcore"
)

type libnutcoreClient interface {
	Info() libnutcore.BackendInfo
	Capabilities() common.CapabilitySet
	GetPermissionSnapshot() (common.PermissionSnapshot, error)
	GetFocusedWindow() (common.FocusedWindowMetadata, error)
	GetFocusedElement() (common.UIElementMetadata, error)
	GetElementAtPoint(position common.Point) (common.UIElementMetadata, error)
	RaiseFocusedWindow() error
	PerformFocusedElementAction(action common.AXAction) error
	PerformElementActionAtPoint(position common.Point, action common.AXAction) error
	FocusElementAtPoint(position common.Point) error
	SearchAXElements(query common.AXElementSearchQuery) ([]common.AXElementMatch, error)
	FocusAXElement(ref common.AXElementRef) error
	PerformAXElementAction(ref common.AXElementRef, action common.AXAction) error
	MoveMouse(position common.Point) error
	GetMousePosition() (common.Point, error)
	MouseClick(button common.MouseButton, double bool) error
	MouseToggle(state common.ButtonState, button common.MouseButton) error
	ScrollMouse(horizontal, vertical int) error
	SetMouseDelay(delay time.Duration) error
	KeyTap(key string, modifiers ...string) error
	KeyToggle(key string, state common.KeyState, modifiers ...string) error
	TypeString(text string) error
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
}
