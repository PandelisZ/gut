package gutmcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PandelisZ/gut/imageproc"
	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/provider"
	"github.com/PandelisZ/gut/shared"
)

type fakeScreenProvider struct {
	size        shared.Region
	image       shared.Image
	regionImage shared.Image
}

func (f *fakeScreenProvider) GrabScreen(context.Context) (shared.Image, error) {
	return f.image, nil
}

func (f *fakeScreenProvider) GrabScreenRegion(context.Context, shared.Region) (shared.Image, error) {
	if f.regionImage.Width != 0 || f.regionImage.Height != 0 {
		return f.regionImage, nil
	}
	return f.image, nil
}

func (f *fakeScreenProvider) HighlightScreenRegion(context.Context, shared.Region, time.Duration, float64) error {
	return nil
}

func (f *fakeScreenProvider) ScreenWidth(context.Context) (int, error) {
	return f.size.Width, nil
}

func (f *fakeScreenProvider) ScreenHeight(context.Context) (int, error) {
	return f.size.Height, nil
}

func (f *fakeScreenProvider) ScreenSize(context.Context) (shared.Region, error) {
	return f.size, nil
}

type fakeWindowProvider struct {
	windows          []shared.WindowHandle
	active           shared.WindowHandle
	titles           map[shared.WindowHandle]string
	regions          map[shared.WindowHandle]shared.Region
	lastFocus        shared.WindowHandle
	lastMoveHandle   shared.WindowHandle
	lastMoveOrigin   shared.Point
	lastResizeHandle shared.WindowHandle
	lastResizeSize   shared.Size
	lastMinimize     shared.WindowHandle
	lastRestore      shared.WindowHandle
}

func (f *fakeWindowProvider) GetWindows(context.Context) ([]shared.WindowHandle, error) {
	return append([]shared.WindowHandle(nil), f.windows...), nil
}

func (f *fakeWindowProvider) GetActiveWindow(context.Context) (shared.WindowHandle, error) {
	return f.active, nil
}

func (f *fakeWindowProvider) GetWindowTitle(_ context.Context, handle shared.WindowHandle) (string, error) {
	return f.titles[handle], nil
}

func (f *fakeWindowProvider) GetWindowRegion(_ context.Context, handle shared.WindowHandle) (shared.Region, error) {
	return f.regions[handle], nil
}

func (f *fakeWindowProvider) FocusWindow(_ context.Context, handle shared.WindowHandle) (bool, error) {
	f.lastFocus = handle
	f.active = handle
	return true, nil
}

func (f *fakeWindowProvider) MoveWindow(_ context.Context, handle shared.WindowHandle, origin shared.Point) (bool, error) {
	f.lastMoveHandle = handle
	f.lastMoveOrigin = origin
	region := f.regions[handle]
	region.Left = origin.X
	region.Top = origin.Y
	f.regions[handle] = region
	return true, nil
}

func (f *fakeWindowProvider) ResizeWindow(_ context.Context, handle shared.WindowHandle, size shared.Size) (bool, error) {
	f.lastResizeHandle = handle
	f.lastResizeSize = size
	region := f.regions[handle]
	region.Width = size.Width
	region.Height = size.Height
	f.regions[handle] = region
	return true, nil
}

func (f *fakeWindowProvider) MinimizeWindow(_ context.Context, handle shared.WindowHandle) (bool, error) {
	f.lastMinimize = handle
	return true, nil
}

func (f *fakeWindowProvider) RestoreWindow(_ context.Context, handle shared.WindowHandle) (bool, error) {
	f.lastRestore = handle
	return true, nil
}

type fakeElementInspectionProvider struct {
	root           shared.WindowElement
	findElements   []shared.WindowElement
	lastHandle     shared.WindowHandle
	lastMax        int
	lastDesc       shared.WindowElementDescription
	getElementsErr error
	findErr        error
}

func (f *fakeElementInspectionProvider) GetElements(context.Context, shared.WindowHandle, int) (shared.WindowElement, error) {
	if f.getElementsErr != nil {
		return shared.WindowElement{}, f.getElementsErr
	}
	return f.root, nil
}

func (f *fakeElementInspectionProvider) FindElement(ctx context.Context, handle shared.WindowHandle, desc shared.WindowElementDescription) (shared.WindowElement, error) {
	elements, err := f.FindElements(ctx, handle, desc)
	if err != nil {
		return shared.WindowElement{}, err
	}
	if len(elements) == 0 {
		return shared.WindowElement{}, errors.New("element not found")
	}
	return elements[0], nil
}

func (f *fakeElementInspectionProvider) FindElements(_ context.Context, handle shared.WindowHandle, desc shared.WindowElementDescription) ([]shared.WindowElement, error) {
	f.lastHandle = handle
	f.lastDesc = desc
	if f.findErr != nil {
		return nil, f.findErr
	}
	return append([]shared.WindowElement(nil), f.findElements...), nil
}

type fakeAccessibilityProvider struct {
	capabilities        common.CapabilitySet
	permissionSnapshot  common.PermissionSnapshot
	permissionErr       error
	focusedWindow       common.FocusedWindowMetadata
	focusedWindowErr    error
	focusedElement      common.UIElementMetadata
	focusedElementErr   error
	searchMatches       []common.AXElementMatch
	searchErr           error
	raisedFocusedWindow bool
	lastFocusedAction   common.AXAction
	lastPointAction     struct {
		point  shared.Point
		action common.AXAction
	}
	lastFocusedPoint shared.Point
	lastFocusedRef   common.AXElementRef
	lastRefAction    struct {
		ref    common.AXElementRef
		action common.AXAction
	}
}

func (f *fakeAccessibilityProvider) GetPermissionSnapshot(context.Context) (common.PermissionSnapshot, error) {
	return f.permissionSnapshot, f.permissionErr
}

func (f *fakeAccessibilityProvider) GetFocusedWindow(context.Context) (common.FocusedWindowMetadata, error) {
	return f.focusedWindow, f.focusedWindowErr
}

func (f *fakeAccessibilityProvider) GetFocusedElement(context.Context) (common.UIElementMetadata, error) {
	return f.focusedElement, f.focusedElementErr
}

func (f *fakeAccessibilityProvider) GetElementAtPoint(context.Context, shared.Point) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, nil
}

func (f *fakeAccessibilityProvider) RaiseFocusedWindow(context.Context) error {
	f.raisedFocusedWindow = true
	return nil
}

func (f *fakeAccessibilityProvider) PerformFocusedElementAction(_ context.Context, action common.AXAction) error {
	f.lastFocusedAction = action
	return nil
}

func (f *fakeAccessibilityProvider) PerformElementActionAtPoint(_ context.Context, point shared.Point, action common.AXAction) error {
	f.lastPointAction.point = point
	f.lastPointAction.action = action
	return nil
}

func (f *fakeAccessibilityProvider) FocusElementAtPoint(_ context.Context, point shared.Point) error {
	f.lastFocusedPoint = point
	return nil
}

func (f *fakeAccessibilityProvider) SearchAXElements(context.Context, common.AXElementSearchQuery) ([]common.AXElementMatch, error) {
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return append([]common.AXElementMatch(nil), f.searchMatches...), nil
}

func (f *fakeAccessibilityProvider) FocusAXElement(_ context.Context, ref common.AXElementRef) error {
	f.lastFocusedRef = ref
	return nil
}

func (f *fakeAccessibilityProvider) PerformAXElementAction(_ context.Context, ref common.AXElementRef, action common.AXAction) error {
	f.lastRefAction.ref = ref
	f.lastRefAction.action = action
	return nil
}

func (f *fakeAccessibilityProvider) Capabilities() common.CapabilitySet {
	return f.capabilities
}

type fakeKeyboardProvider struct {
	typed    []string
	clicked  []shared.Key
	pressed  []shared.Key
	released []shared.Key
}

func (f *fakeKeyboardProvider) SetKeyboardDelay(time.Duration) {}
func (f *fakeKeyboardProvider) Type(_ context.Context, input string) error {
	f.typed = append(f.typed, input)
	return nil
}
func (f *fakeKeyboardProvider) Click(_ context.Context, keys ...shared.Key) error {
	f.clicked = append([]shared.Key(nil), keys...)
	return nil
}
func (f *fakeKeyboardProvider) PressKey(_ context.Context, keys ...shared.Key) error {
	f.pressed = append([]shared.Key(nil), keys...)
	return nil
}
func (f *fakeKeyboardProvider) ReleaseKey(_ context.Context, keys ...shared.Key) error {
	f.released = append([]shared.Key(nil), keys...)
	return nil
}

type fakeMouseProvider struct {
	position    shared.Point
	lastClick   shared.Button
	lastDouble  shared.Button
	lastPress   shared.Button
	lastRelease shared.Button
	scroll      struct {
		up    int
		down  int
		left  int
		right int
	}
}

func (f *fakeMouseProvider) SetMouseDelay(time.Duration) {}
func (f *fakeMouseProvider) SetMousePosition(_ context.Context, point shared.Point) error {
	f.position = point
	return nil
}
func (f *fakeMouseProvider) CurrentMousePosition(context.Context) (shared.Point, error) {
	return f.position, nil
}
func (f *fakeMouseProvider) Click(_ context.Context, button shared.Button) error {
	f.lastClick = button
	return nil
}
func (f *fakeMouseProvider) DoubleClick(_ context.Context, button shared.Button) error {
	f.lastDouble = button
	return nil
}
func (f *fakeMouseProvider) ScrollUp(_ context.Context, amount int) error {
	f.scroll.up += amount
	return nil
}
func (f *fakeMouseProvider) ScrollDown(_ context.Context, amount int) error {
	f.scroll.down += amount
	return nil
}
func (f *fakeMouseProvider) ScrollLeft(_ context.Context, amount int) error {
	f.scroll.left += amount
	return nil
}
func (f *fakeMouseProvider) ScrollRight(_ context.Context, amount int) error {
	f.scroll.right += amount
	return nil
}
func (f *fakeMouseProvider) PressButton(_ context.Context, button shared.Button) error {
	f.lastPress = button
	return nil
}
func (f *fakeMouseProvider) ReleaseButton(_ context.Context, button shared.Button) error {
	f.lastRelease = button
	return nil
}

type fakeClipboardProvider struct {
	text string
}

func (f *fakeClipboardProvider) HasText(context.Context) (bool, error) {
	return len(f.text) != 0, nil
}

func (f *fakeClipboardProvider) Clear(context.Context) (bool, error) {
	f.text = ""
	return true, nil
}

func (f *fakeClipboardProvider) Copy(_ context.Context, text string) error {
	f.text = text
	return nil
}

func (f *fakeClipboardProvider) Paste(context.Context) (string, error) {
	return f.text, nil
}

type testDeps struct {
	service       *Service
	registry      *provider.Registry
	screen        *fakeScreenProvider
	windows       *fakeWindowProvider
	inspection    *fakeElementInspectionProvider
	accessibility *fakeAccessibilityProvider
	keyboard      *fakeKeyboardProvider
	mouse         *fakeMouseProvider
	clipboard     *fakeClipboardProvider
}

func newTestDeps(t *testing.T, allowMutation bool) testDeps {
	t.Helper()

	screenImage := shared.Image{
		Width:     2,
		Height:    2,
		Data:      []byte{255, 0, 0, 255, 0, 255, 0, 255, 255, 0, 0, 255, 0, 0, 255, 255},
		Channels:  4,
		ByteWidth: 8,
		ColorMode: shared.ColorModeRGB,
		PixelDensity: shared.PixelDensity{
			ScaleX: 1,
			ScaleY: 1,
		},
	}

	screen := &fakeScreenProvider{
		size:        shared.Region{Left: 0, Top: 0, Width: 2, Height: 2},
		image:       screenImage,
		regionImage: screenImage,
	}
	windows := &fakeWindowProvider{
		windows: []shared.WindowHandle{1, 2},
		active:  2,
		titles: map[shared.WindowHandle]string{
			1: "Editor",
			2: "Terminal",
		},
		regions: map[shared.WindowHandle]shared.Region{
			1: {Left: 10, Top: 20, Width: 300, Height: 200},
			2: {Left: 40, Top: 60, Width: 500, Height: 400},
		},
	}
	inspection := &fakeElementInspectionProvider{
		root: shared.WindowElement{
			Role:  stringPtr("window"),
			Title: stringPtr("Terminal"),
			Children: []shared.WindowElement{
				{Role: stringPtr("button"), Title: stringPtr("Run")},
			},
		},
		findElements: []shared.WindowElement{
			{Role: stringPtr("button"), Title: stringPtr("Run")},
			{Role: stringPtr("button"), Title: stringPtr("Stop")},
		},
	}
	accessibility := &fakeAccessibilityProvider{
		capabilities: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityWindowList, Availability: common.AvailabilityAvailable},
			common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityAvailable},
			common.CapabilityStatus{Capability: common.CapabilityAXElementSearch, Availability: common.AvailabilityAvailable},
		),
		permissionSnapshot: common.PermissionSnapshot{
			Accessibility:   common.PermissionStatus{Granted: true, Supported: true},
			ScreenRecording: common.PermissionStatus{Granted: true, Supported: true},
		},
		focusedWindow: common.FocusedWindowMetadata{
			Handle:    2,
			Title:     "Terminal",
			Role:      "AXWindow",
			Subrole:   "AXStandardWindow",
			Rect:      common.Rect{X: 40, Y: 60, Width: 500, Height: 400},
			RectKnown: true,
			Focused:   true,
			Main:      true,
			OwnerPID:  123,
			OwnerName: "Terminal",
			BundleID:  "com.apple.Terminal",
		},
		focusedElement: common.UIElementMetadata{
			Role:       "AXTextField",
			Title:      "Command",
			Value:      "echo hi",
			Enabled:    true,
			Focused:    true,
			Frame:      common.Rect{X: 80, Y: 120, Width: 200, Height: 30},
			FrameKnown: true,
			Actions:    []string{"AXConfirm"},
		},
		searchMatches: []common.AXElementMatch{
			{
				Ref: common.AXElementRef{Scope: common.AXSearchScopeWindowHandle, WindowHandle: 2, Path: []int{0, 1}},
				Metadata: common.UIElementMetadata{
					Role:    "AXButton",
					Title:   "Run",
					Enabled: true,
				},
				Depth:            1,
				ActionPointKnown: true,
				ActionPoint:      common.Point{X: 90, Y: 90},
			},
		},
	}
	keyboard := &fakeKeyboardProvider{}
	mouse := &fakeMouseProvider{position: shared.Point{X: 5, Y: 5}}
	clipboard := &fakeClipboardProvider{text: "seed"}

	registry := provider.NewRegistry()
	registry.RegisterScreen(screen)
	registry.RegisterWindow(windows)
	registry.RegisterElementInspection(inspection)
	registry.RegisterAccessibility(accessibility)
	registry.RegisterKeyboard(keyboard)
	registry.RegisterMouse(mouse)
	registry.RegisterClipboard(clipboard)
	registry.RegisterImageProcessor(imageproc.NewProcessor())
	registry.RegisterColorFinder(imageproc.NewColorFinder(imageproc.NewProcessor()))
	registry.RegisterWindowFinder(imageproc.NewWindowFinder(windows))

	return testDeps{
		service:       NewService(registry, Config{AllowMutation: allowMutation}),
		registry:      registry,
		screen:        screen,
		windows:       windows,
		inspection:    inspection,
		accessibility: accessibility,
		keyboard:      keyboard,
		mouse:         mouse,
		clipboard:     clipboard,
	}
}

func stringPtr(value string) *string {
	return &value
}
