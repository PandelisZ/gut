package provider

import (
	"context"
	"time"

	gutlog "gut/log"
	"gut/native/common"
	"gut/shared"
)

type DataSource[Parameter any, Result any] interface {
	Load(ctx context.Context, parameter Parameter) (Result, error)
}

type DataSink[Parameter any, Result any] interface {
	Store(ctx context.Context, parameter Parameter) (Result, error)
}

type KeyboardProvider interface {
	SetKeyboardDelay(delay time.Duration)
	Type(ctx context.Context, input string) error
	Click(ctx context.Context, keys ...shared.Key) error
	PressKey(ctx context.Context, keys ...shared.Key) error
	ReleaseKey(ctx context.Context, keys ...shared.Key) error
}

type MouseProvider interface {
	SetMouseDelay(delay time.Duration)
	SetMousePosition(ctx context.Context, point shared.Point) error
	CurrentMousePosition(ctx context.Context) (shared.Point, error)
	Click(ctx context.Context, button shared.Button) error
	DoubleClick(ctx context.Context, button shared.Button) error
	ScrollUp(ctx context.Context, amount int) error
	ScrollDown(ctx context.Context, amount int) error
	ScrollLeft(ctx context.Context, amount int) error
	ScrollRight(ctx context.Context, amount int) error
	PressButton(ctx context.Context, button shared.Button) error
	ReleaseButton(ctx context.Context, button shared.Button) error
}

type ScreenProvider interface {
	GrabScreen(ctx context.Context) (shared.Image, error)
	GrabScreenRegion(ctx context.Context, region shared.Region) (shared.Image, error)
	HighlightScreenRegion(ctx context.Context, region shared.Region, duration time.Duration, opacity float64) error
	ScreenWidth(ctx context.Context) (int, error)
	ScreenHeight(ctx context.Context) (int, error)
	ScreenSize(ctx context.Context) (shared.Region, error)
}

type WindowProvider interface {
	GetWindows(ctx context.Context) ([]shared.WindowHandle, error)
	GetActiveWindow(ctx context.Context) (shared.WindowHandle, error)
	GetWindowTitle(ctx context.Context, windowHandle shared.WindowHandle) (string, error)
	GetWindowRegion(ctx context.Context, windowHandle shared.WindowHandle) (shared.Region, error)
	FocusWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error)
	MoveWindow(ctx context.Context, windowHandle shared.WindowHandle, newOrigin shared.Point) (bool, error)
	ResizeWindow(ctx context.Context, windowHandle shared.WindowHandle, newSize shared.Size) (bool, error)
	MinimizeWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error)
	RestoreWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error)
}

type AccessibilityProvider interface {
	GetPermissionSnapshot(ctx context.Context) (common.PermissionSnapshot, error)
	GetFocusedWindow(ctx context.Context) (common.FocusedWindowMetadata, error)
	GetFocusedElement(ctx context.Context) (common.UIElementMetadata, error)
	GetElementAtPoint(ctx context.Context, point shared.Point) (common.UIElementMetadata, error)
	RaiseFocusedWindow(ctx context.Context) error
	PerformFocusedElementAction(ctx context.Context, action common.AXAction) error
	PerformElementActionAtPoint(ctx context.Context, point shared.Point, action common.AXAction) error
	FocusElementAtPoint(ctx context.Context, point shared.Point) error
	SearchAXElements(ctx context.Context, query common.AXElementSearchQuery) ([]common.AXElementMatch, error)
	FocusAXElement(ctx context.Context, ref common.AXElementRef) error
	PerformAXElementAction(ctx context.Context, ref common.AXElementRef, action common.AXAction) error
	Capabilities() common.CapabilitySet
}

type ImageFinder interface {
	FindMatch(ctx context.Context, request shared.MatchRequest[shared.Image, any]) (shared.MatchResult[shared.Region], error)
	FindMatches(ctx context.Context, request shared.MatchRequest[shared.Image, any]) ([]shared.MatchResult[shared.Region], error)
}

type ImageReader interface {
	Load(ctx context.Context, path string) (shared.Image, error)
}

type ImageWriterParameters struct {
	Image shared.Image
	Path  string
}

type ImageWriter interface {
	Store(ctx context.Context, parameters ImageWriterParameters) error
}

type ImageProcessor interface {
	ColorAt(ctx context.Context, image shared.Image, location shared.Point) (shared.RGBA, error)
}

type TextFinder interface {
	FindMatch(ctx context.Context, request shared.MatchRequest[shared.TextQuery, any]) (shared.MatchResult[shared.Region], error)
	FindMatches(ctx context.Context, request shared.MatchRequest[shared.TextQuery, any]) ([]shared.MatchResult[shared.Region], error)
}

type WindowFinder interface {
	FindMatch(ctx context.Context, query shared.WindowQuery) (shared.WindowHandle, error)
	FindMatches(ctx context.Context, query shared.WindowQuery) ([]shared.WindowHandle, error)
}

type ColorFinder interface {
	FindMatch(ctx context.Context, request shared.MatchRequest[shared.ColorQuery, any]) (shared.MatchResult[shared.Point], error)
	FindMatches(ctx context.Context, request shared.MatchRequest[shared.ColorQuery, any]) ([]shared.MatchResult[shared.Point], error)
}

type ElementInspectionProvider interface {
	GetElements(ctx context.Context, windowHandle shared.WindowHandle, maxElements int) (shared.WindowElement, error)
	FindElement(ctx context.Context, windowHandle shared.WindowHandle, description shared.WindowElementDescription) (shared.WindowElement, error)
	FindElements(ctx context.Context, windowHandle shared.WindowHandle, description shared.WindowElementDescription) ([]shared.WindowElement, error)
}

type ClipboardProvider interface {
	HasText(ctx context.Context) (bool, error)
	Clear(ctx context.Context) (bool, error)
	Copy(ctx context.Context, text string) error
	Paste(ctx context.Context) (string, error)
}

type LogProvider = gutlog.Logger
