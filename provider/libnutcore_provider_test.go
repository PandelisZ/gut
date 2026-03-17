package provider

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"strings"
	"testing"
	"time"

	"gut/native/common"
	"gut/native/libnutcore"
	"gut/shared"
)

type recordedToggle struct {
	key   string
	state common.KeyState
}

type fakeLibnutcoreClient struct {
	info                         libnutcore.BackendInfo
	capabilities                 common.CapabilitySet
	permissionSnapshot           common.PermissionSnapshot
	permissionSnapshotErr        error
	focusedWindow                common.FocusedWindowMetadata
	focusedWindowErr             error
	focusedElement               common.UIElementMetadata
	focusedElementErr            error
	elementAtPointPosition       common.Point
	elementAtPoint               common.UIElementMetadata
	elementAtPointErr            error
	raiseFocusedWindowErr        error
	performFocusedElementAction  common.AXAction
	performFocusedElementErr     error
	elementActionAtPointPosition common.Point
	elementActionAtPointAction   common.AXAction
	elementActionAtPointErr      error
	focusElementAtPointPosition  common.Point
	focusElementAtPointErr       error
	keyTapKey                    string
	keyTapModifiers              []string
	keyTapErr                    error
	typedText                    []string
	typeStringErr                error
	keyToggles                   []recordedToggle
	keyToggleErr                 error
	mouseClickButton             common.MouseButton
	mouseClickDouble             bool
	mouseClicks                  []common.MouseButton
	mouseClickErr                error
	mouseToggleStates            []common.ButtonState
	mouseToggleButton            []common.MouseButton
	mouseToggleErr               error
	scrolls                      [][2]int
	scrollErr                    error
	moveMousePoint               common.Point
	moveMouseErr                 error
	mousePosition                common.Point
	mousePositionErr             error
	captureRegion                *common.Rect
	captureBitmap                *common.Bitmap
	captureErr                   error
	screenSize                   common.Size
	screenSizeErr                error
	highlightCalls               int
	captureCalls                 int
	highlightRegion              common.Rect
	highlightDuration            time.Duration
	highlightOpacity             float64
	highlightErr                 error
	windows                      []common.WindowHandle
	getWindowsErr                error
	activeWindow                 common.WindowHandle
	activeWindowErr              error
	windowRectHandle             common.WindowHandle
	windowRect                   common.Rect
	windowRectErr                error
	windowTitleHandle            common.WindowHandle
	windowTitle                  string
	windowTitleErr               error
	focusHandle                  common.WindowHandle
	focusResult                  bool
	focusErr                     error
	resizeHandle                 common.WindowHandle
	resizeSize                   common.Size
	resizeResult                 bool
	resizeErr                    error
	moveWindowHandle             common.WindowHandle
	moveWindowOrigin             common.Point
	moveWindowResult             bool
	moveWindowErr                error
	unavailableMode              bool
	minimizeHandle               common.WindowHandle
	minimizeErr                  error
	restoreHandle                common.WindowHandle
	restoreErr                   error
}

func (f *fakeLibnutcoreClient) Info() libnutcore.BackendInfo { return f.info }
func (f *fakeLibnutcoreClient) Capabilities() common.CapabilitySet {
	if f.capabilities == nil {
		return common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityAvailable},
			common.CapabilityStatus{Capability: common.CapabilityScreenHighlight, Availability: common.AvailabilityAvailable},
		)
	}
	return f.capabilities
}
func (f *fakeLibnutcoreClient) GetPermissionSnapshot() (common.PermissionSnapshot, error) {
	return f.permissionSnapshot, f.permissionSnapshotErr
}
func (f *fakeLibnutcoreClient) GetFocusedWindow() (common.FocusedWindowMetadata, error) {
	return f.focusedWindow, f.focusedWindowErr
}
func (f *fakeLibnutcoreClient) GetFocusedElement() (common.UIElementMetadata, error) {
	return f.focusedElement, f.focusedElementErr
}
func (f *fakeLibnutcoreClient) GetElementAtPoint(position common.Point) (common.UIElementMetadata, error) {
	f.elementAtPointPosition = position
	return f.elementAtPoint, f.elementAtPointErr
}
func (f *fakeLibnutcoreClient) RaiseFocusedWindow() error {
	return f.raiseFocusedWindowErr
}
func (f *fakeLibnutcoreClient) PerformFocusedElementAction(action common.AXAction) error {
	f.performFocusedElementAction = action
	return f.performFocusedElementErr
}
func (f *fakeLibnutcoreClient) PerformElementActionAtPoint(position common.Point, action common.AXAction) error {
	f.elementActionAtPointPosition = position
	f.elementActionAtPointAction = action
	return f.elementActionAtPointErr
}
func (f *fakeLibnutcoreClient) FocusElementAtPoint(position common.Point) error {
	f.focusElementAtPointPosition = position
	return f.focusElementAtPointErr
}
func (f *fakeLibnutcoreClient) DragMouse(position common.Point, button common.MouseButton) error {
	return nil
}
func (f *fakeLibnutcoreClient) MoveMouse(position common.Point) error {
	f.moveMousePoint = position
	return f.moveMouseErr
}
func (f *fakeLibnutcoreClient) GetMousePosition() (common.Point, error) {
	return f.mousePosition, f.mousePositionErr
}
func (f *fakeLibnutcoreClient) MouseClick(button common.MouseButton, double bool) error {
	f.mouseClickButton = button
	f.mouseClickDouble = double
	f.mouseClicks = append(f.mouseClicks, button)
	return f.mouseClickErr
}
func (f *fakeLibnutcoreClient) MouseToggle(state common.ButtonState, button common.MouseButton) error {
	f.mouseToggleStates = append(f.mouseToggleStates, state)
	f.mouseToggleButton = append(f.mouseToggleButton, button)
	return f.mouseToggleErr
}
func (f *fakeLibnutcoreClient) ScrollMouse(horizontal, vertical int) error {
	f.scrolls = append(f.scrolls, [2]int{horizontal, vertical})
	if f.scrollErr != nil {
		return f.scrollErr
	}
	if f.unavailableMode {
		return common.UnavailableOperation("scrollMouse", "linux", common.CapabilityMouseScroll, "fake native backend unavailable")
	}
	return nil
}
func (f *fakeLibnutcoreClient) SetMouseDelay(delay time.Duration) error { return nil }
func (f *fakeLibnutcoreClient) KeyTap(key string, modifiers ...string) error {
	f.keyTapKey = key
	f.keyTapModifiers = append([]string(nil), modifiers...)
	if f.keyTapErr != nil {
		return f.keyTapErr
	}
	if f.unavailableMode {
		return common.UnavailableOperation("keyTap", "linux", common.CapabilityKeyboardTap, "fake native backend unavailable")
	}
	return nil
}
func (f *fakeLibnutcoreClient) KeyToggle(key string, state common.KeyState, modifiers ...string) error {
	f.keyToggles = append(f.keyToggles, recordedToggle{key: key, state: state})
	if f.unavailableMode {
		return common.UnavailableOperation("keyToggle", "linux", common.CapabilityKeyboardToggle, "fake native backend unavailable")
	}
	return f.keyToggleErr
}
func (f *fakeLibnutcoreClient) TypeString(text string) error {
	f.typedText = append(f.typedText, text)
	return f.typeStringErr
}
func (f *fakeLibnutcoreClient) TypeStringDelayed(text string, cpm int) error { return nil }
func (f *fakeLibnutcoreClient) SetKeyboardDelay(delay time.Duration) error   { return nil }
func (f *fakeLibnutcoreClient) GetScreenSize() (common.Size, error) {
	if f.screenSizeErr != nil {
		return f.screenSize, f.screenSizeErr
	}
	if f.unavailableMode {
		return common.Size{}, common.UnavailableOperation("getScreenSize", "linux", common.CapabilityScreenSize, "fake native backend unavailable")
	}
	return f.screenSize, nil
}
func (f *fakeLibnutcoreClient) Highlight(region common.Rect, duration time.Duration, opacity float64) error {
	f.highlightCalls++
	f.highlightRegion = region
	f.highlightDuration = duration
	f.highlightOpacity = opacity
	return f.highlightErr
}
func (f *fakeLibnutcoreClient) CaptureScreen(region *common.Rect) (*common.Bitmap, error) {
	f.captureCalls++
	if region != nil {
		copy := *region
		f.captureRegion = &copy
	} else {
		f.captureRegion = nil
	}
	return f.captureBitmap, f.captureErr
}
func (f *fakeLibnutcoreClient) GetWindows() ([]common.WindowHandle, error) {
	return append([]common.WindowHandle(nil), f.windows...), f.getWindowsErr
}
func (f *fakeLibnutcoreClient) GetActiveWindow() (common.WindowHandle, error) {
	return f.activeWindow, f.activeWindowErr
}
func (f *fakeLibnutcoreClient) GetWindowRect(handle common.WindowHandle) (common.Rect, error) {
	f.windowRectHandle = handle
	return f.windowRect, f.windowRectErr
}
func (f *fakeLibnutcoreClient) GetWindowTitle(handle common.WindowHandle) (string, error) {
	f.windowTitleHandle = handle
	return f.windowTitle, f.windowTitleErr
}
func (f *fakeLibnutcoreClient) FocusWindow(handle common.WindowHandle) (bool, error) {
	f.focusHandle = handle
	return f.focusResult, f.focusErr
}
func (f *fakeLibnutcoreClient) ResizeWindow(handle common.WindowHandle, size common.Size) (bool, error) {
	f.resizeHandle = handle
	f.resizeSize = size
	return f.resizeResult, f.resizeErr
}
func (f *fakeLibnutcoreClient) MoveWindow(handle common.WindowHandle, origin common.Point) (bool, error) {
	f.moveWindowHandle = handle
	f.moveWindowOrigin = origin
	return f.moveWindowResult, f.moveWindowErr
}
func (f *fakeLibnutcoreClient) MinimizeWindow(handle common.WindowHandle) (bool, error) {
	f.minimizeHandle = handle
	if f.minimizeErr != nil {
		return false, f.minimizeErr
	}
	if f.unavailableMode {
		return false, common.CapabilityUnavailable("minimizeWindow", "linux", common.CapabilityWindowMinimize, "fake native backend unavailable")
	}
	return false, nil
}
func (f *fakeLibnutcoreClient) RestoreWindow(handle common.WindowHandle) (bool, error) {
	f.restoreHandle = handle
	return false, f.restoreErr
}
func (f *fakeLibnutcoreClient) GetXDisplayName() (string, error)  { return "", nil }
func (f *fakeLibnutcoreClient) SetXDisplayName(name string) error { return nil }

func testPNGBytes(width, height int) []byte {
	buffer := bytes.Buffer{}
	imageData := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			imageData.SetRGBA(x, y, color.RGBA{R: uint8(x + 1), G: uint8(y + 1), B: 0x7f, A: 0xff})
		}
	}
	if err := png.Encode(&buffer, imageData); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func withScreenFallbackTestHooks(t *testing.T, goos string, screencapture func(context.Context, *common.Rect) ([]byte, error)) {
	t.Helper()
	previousGOOS := libnutcoreScreenGOOS
	previousScreencapture := libnutcoreMacOSScreencapture
	libnutcoreScreenGOOS = goos
	libnutcoreMacOSScreencapture = screencapture
	t.Cleanup(func() {
		libnutcoreScreenGOOS = previousGOOS
		libnutcoreMacOSScreencapture = previousScreencapture
	})
}

func TestAccessibilityProviderForwardsMetadataAndCapabilities(t *testing.T) {
	capabilities := common.NewCapabilitySet(
		common.CapabilityStatus{Capability: common.CapabilityPermissionReadiness, Availability: common.AvailabilityPermissionBlocked, Reason: "accessibility permission denied"},
		common.CapabilityStatus{Capability: common.CapabilityAXFocusedWindowMetadata, Availability: common.AvailabilityAvailable},
		common.CapabilityStatus{Capability: common.CapabilityAXFocusedElementMetadata, Availability: common.AvailabilityAvailable},
		common.CapabilityStatus{Capability: common.CapabilityAXElementAtPointMetadata, Availability: common.AvailabilityUnsupported, Reason: "element lookup not implemented"},
	)
	permissionSnapshot := common.PermissionSnapshot{
		Accessibility:   common.PermissionStatus{Granted: false, Supported: true, Reason: "accessibility permission denied"},
		ScreenRecording: common.PermissionStatus{Granted: true, Supported: true},
	}
	focusedWindow := common.FocusedWindowMetadata{
		Handle:    77,
		Title:     "Editor",
		Role:      "AXWindow",
		Subrole:   "AXStandardWindow",
		Rect:      common.Rect{X: 10, Y: 20, Width: 300, Height: 200},
		RectKnown: true,
		Focused:   true,
		Main:      true,
		OwnerPID:  42,
		OwnerName: "TestApp",
		BundleID:  "com.example.test",
	}
	focusedElement := common.UIElementMetadata{
		Role:        "AXTextField",
		Subrole:     "AXSearchField",
		Title:       "Search",
		Description: "Search field",
		Value:       "query",
		Enabled:     true,
		Focused:     true,
		Frame:       common.Rect{X: 30, Y: 40, Width: 120, Height: 24},
		FrameKnown:  true,
		Actions:     []string{"AXPress", "AXConfirm"},
	}
	pointElement := common.UIElementMetadata{
		Role:        "AXButton",
		Title:       "Go",
		Description: "Run search",
		Enabled:     true,
		Frame:       common.Rect{X: 160, Y: 40, Width: 44, Height: 24},
		FrameKnown:  true,
	}
	client := &fakeLibnutcoreClient{
		capabilities:       capabilities,
		permissionSnapshot: permissionSnapshot,
		focusedWindow:      focusedWindow,
		focusedElement:     focusedElement,
		elementAtPoint:     pointElement,
	}
	provider := NewLibnutcoreAccessibilityProvider(client)

	gotPermissionSnapshot, err := provider.GetPermissionSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected permission snapshot error: %v", err)
	}
	if !reflect.DeepEqual(gotPermissionSnapshot, permissionSnapshot) {
		t.Fatalf("unexpected permission snapshot: %#v", gotPermissionSnapshot)
	}

	gotFocusedWindow, err := provider.GetFocusedWindow(context.Background())
	if err != nil {
		t.Fatalf("unexpected focused window error: %v", err)
	}
	if !reflect.DeepEqual(gotFocusedWindow, focusedWindow) {
		t.Fatalf("unexpected focused window metadata: %#v", gotFocusedWindow)
	}

	gotFocusedElement, err := provider.GetFocusedElement(context.Background())
	if err != nil {
		t.Fatalf("unexpected focused element error: %v", err)
	}
	if !reflect.DeepEqual(gotFocusedElement, focusedElement) {
		t.Fatalf("unexpected focused element metadata: %#v", gotFocusedElement)
	}

	gotPointElement, err := provider.GetElementAtPoint(context.Background(), shared.Point{X: 160, Y: 52})
	if err != nil {
		t.Fatalf("unexpected element-at-point error: %v", err)
	}
	if client.elementAtPointPosition != (common.Point{X: 160, Y: 52}) {
		t.Fatalf("unexpected element-at-point position: %#v", client.elementAtPointPosition)
	}
	if !reflect.DeepEqual(gotPointElement, pointElement) {
		t.Fatalf("unexpected element-at-point metadata: %#v", gotPointElement)
	}
	if !reflect.DeepEqual(provider.Capabilities(), capabilities) {
		t.Fatalf("unexpected capabilities: %#v", provider.Capabilities())
	}
}

func TestAccessibilityProviderForwardsActions(t *testing.T) {
	actionErr := errors.New("ax action failed")
	client := &fakeLibnutcoreClient{
		raiseFocusedWindowErr: actionErr,
	}
	provider := NewLibnutcoreAccessibilityProvider(client)

	if err := provider.RaiseFocusedWindow(context.Background()); !errors.Is(err, actionErr) {
		t.Fatalf("expected raise-focused-window error, got %v", err)
	}

	client.raiseFocusedWindowErr = nil
	client.performFocusedElementErr = actionErr
	if err := provider.PerformFocusedElementAction(context.Background(), common.AXConfirm); !errors.Is(err, actionErr) {
		t.Fatalf("expected focused-element action error, got %v", err)
	}
	if client.performFocusedElementAction != common.AXConfirm {
		t.Fatalf("unexpected focused-element action forwarding: %q", client.performFocusedElementAction)
	}

	client.performFocusedElementErr = nil
	client.elementActionAtPointErr = actionErr
	if err := provider.PerformElementActionAtPoint(context.Background(), shared.Point{X: 45, Y: 67}, common.AXShowMenu); !errors.Is(err, actionErr) {
		t.Fatalf("expected element-at-point action error, got %v", err)
	}
	if client.elementActionAtPointPosition != (common.Point{X: 45, Y: 67}) {
		t.Fatalf("unexpected element-at-point position: %#v", client.elementActionAtPointPosition)
	}
	if client.elementActionAtPointAction != common.AXShowMenu {
		t.Fatalf("unexpected element-at-point action forwarding: %q", client.elementActionAtPointAction)
	}

	client.elementActionAtPointErr = nil
	client.focusElementAtPointErr = actionErr
	if err := provider.FocusElementAtPoint(context.Background(), shared.Point{X: 12, Y: 34}); !errors.Is(err, actionErr) {
		t.Fatalf("expected focus-element-at-point error, got %v", err)
	}
	if client.focusElementAtPointPosition != (common.Point{X: 12, Y: 34}) {
		t.Fatalf("unexpected focus-element-at-point position: %#v", client.focusElementAtPointPosition)
	}
}

func TestAccessibilityProviderChecksContextBeforeWork(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreAccessibilityProvider(client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := provider.GetPermissionSnapshot(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for permission snapshot, got %v", err)
	}
	if _, err := provider.GetFocusedWindow(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for focused window, got %v", err)
	}
	if _, err := provider.GetFocusedElement(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for focused element, got %v", err)
	}
	if _, err := provider.GetElementAtPoint(ctx, shared.Point{X: 1, Y: 2}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for element-at-point, got %v", err)
	}
	if err := provider.RaiseFocusedWindow(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for raise-focused-window, got %v", err)
	}
	if err := provider.PerformFocusedElementAction(ctx, common.AXPress); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for focused-element action, got %v", err)
	}
	if err := provider.PerformElementActionAtPoint(ctx, shared.Point{X: 3, Y: 4}, common.AXPick); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for element-at-point action, got %v", err)
	}
	if err := provider.FocusElementAtPoint(ctx, shared.Point{X: 5, Y: 6}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation for focus-element-at-point, got %v", err)
	}
	if client.elementAtPointPosition != (common.Point{}) {
		t.Fatalf("expected canceled element-at-point lookup to avoid native calls, got %#v", client.elementAtPointPosition)
	}
	if client.performFocusedElementAction != "" {
		t.Fatalf("expected canceled focused-element action to avoid native calls, got %q", client.performFocusedElementAction)
	}
	if client.elementActionAtPointPosition != (common.Point{}) || client.elementActionAtPointAction != "" {
		t.Fatalf("expected canceled element-at-point action to avoid native calls, got position=%#v action=%q", client.elementActionAtPointPosition, client.elementActionAtPointAction)
	}
	if client.focusElementAtPointPosition != (common.Point{}) {
		t.Fatalf("expected canceled focus-element-at-point to avoid native calls, got %#v", client.focusElementAtPointPosition)
	}
}

func TestKeyToLibnutTokenUsesMainCCTokens(t *testing.T) {
	cases := map[shared.Key]string{
		shared.KeyLeftSuper:  "meta",
		shared.KeyRightSuper: "right_meta",
		shared.KeyNumPad0:    "numpad_0",
		shared.KeyAudioVolUp: "audio_vol_up",
	}
	for key, want := range cases {
		got, err := keyToLibnutToken(key)
		if err != nil {
			t.Fatalf("unexpected translation error for %s: %v", key, err)
		}
		if got != want {
			t.Fatalf("unexpected token for %s: got %q want %q", key, got, want)
		}
	}
}

func TestKeyboardProviderClickUsesLastKeyAsPrimaryAndPreviousAsModifiers(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreKeyboardProvider(client)

	if err := provider.Click(context.Background(), shared.KeyLeftControl, shared.KeyRightAlt, shared.KeyA); err != nil {
		t.Fatalf("unexpected click error: %v", err)
	}
	if client.keyTapKey != "" || len(client.keyTapModifiers) != 0 {
		t.Fatalf("expected click to avoid native key tap, got key=%q modifiers=%#v", client.keyTapKey, client.keyTapModifiers)
	}
	want := []recordedToggle{
		{key: "control", state: common.KeyStateDown},
		{key: "right_alt", state: common.KeyStateDown},
		{key: "a", state: common.KeyStateDown},
		{key: "a", state: common.KeyStateUp},
		{key: "right_alt", state: common.KeyStateUp},
		{key: "control", state: common.KeyStateUp},
	}
	if !reflect.DeepEqual(client.keyToggles, want) {
		t.Fatalf("unexpected toggle sequence: %#v", client.keyToggles)
	}
}

func TestKeyboardProviderPressAndReleaseOrder(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreKeyboardProvider(client)

	if err := provider.PressKey(context.Background(), shared.KeyLeftControl, shared.KeyA, shared.KeyRightAlt); err != nil {
		t.Fatalf("unexpected press error: %v", err)
	}
	if err := provider.ReleaseKey(context.Background(), shared.KeyLeftControl, shared.KeyA, shared.KeyRightAlt); err != nil {
		t.Fatalf("unexpected release error: %v", err)
	}

	want := []recordedToggle{
		{key: "control", state: common.KeyStateDown},
		{key: "a", state: common.KeyStateDown},
		{key: "right_alt", state: common.KeyStateDown},
		{key: "right_alt", state: common.KeyStateUp},
		{key: "a", state: common.KeyStateUp},
		{key: "control", state: common.KeyStateUp},
	}
	if !reflect.DeepEqual(client.keyToggles, want) {
		t.Fatalf("unexpected toggle sequence: %#v", client.keyToggles)
	}
}

func TestKeyboardProviderChecksContextBeforeWork(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreKeyboardProvider(client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := provider.Click(ctx, shared.KeyA); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if client.keyTapKey != "" || len(client.keyToggles) != 0 {
		t.Fatal("expected no native calls after context cancellation")
	}
}

func TestKeyboardProviderTypeDelegatesAndPropagatesErrors(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreKeyboardProvider(client)

	if err := provider.Type(context.Background(), "hello"); err != nil {
		t.Fatalf("unexpected type error: %v", err)
	}
	if !reflect.DeepEqual(client.typedText, []string{"hello"}) {
		t.Fatalf("unexpected typed text: %#v", client.typedText)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := provider.Type(ctx, "blocked"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if !reflect.DeepEqual(client.typedText, []string{"hello"}) {
		t.Fatalf("expected canceled type to avoid native calls, got %#v", client.typedText)
	}

	typeErr := errors.New("type failed")
	client.typeStringErr = typeErr
	if err := provider.Type(context.Background(), "boom"); !errors.Is(err, typeErr) {
		t.Fatalf("expected native type error, got %v", err)
	}
}

func TestKeyboardProviderClickRejectsUnknownKeysAndUsesDeleteToken(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreKeyboardProvider(client)

	if err := provider.Click(context.Background(), shared.KeyDelete); err != nil {
		t.Fatalf("unexpected delete click error: %v", err)
	}
	if client.keyTapKey != "" {
		t.Fatalf("expected click to avoid native key tap, got %q", client.keyTapKey)
	}
	wantDelete := []recordedToggle{
		{key: "delete", state: common.KeyStateDown},
		{key: "delete", state: common.KeyStateUp},
	}
	if !reflect.DeepEqual(client.keyToggles, wantDelete) {
		t.Fatalf("unexpected delete toggle sequence: %#v", client.keyToggles)
	}

	if err := provider.Click(context.Background(), shared.KeyLeftControl, shared.Key(-1)); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid primary key error, got %v", err)
	}
	if err := provider.Click(context.Background(), shared.Key(-1), shared.KeyA); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid modifier key error, got %v", err)
	}
}

func TestKeyboardProviderPropagatesNativeToggleAndTapErrors(t *testing.T) {
	t.Run("click", func(t *testing.T) {
		toggleErr := errors.New("toggle failed")
		client := &fakeLibnutcoreClient{keyToggleErr: toggleErr}
		provider := NewLibnutcoreKeyboardProvider(client)

		if err := provider.Click(context.Background(), shared.KeyLeftShift, shared.KeyA); !errors.Is(err, toggleErr) {
			t.Fatalf("expected toggle error, got %v", err)
		}
	})

	t.Run("press", func(t *testing.T) {
		toggleErr := errors.New("toggle failed")
		client := &fakeLibnutcoreClient{keyToggleErr: toggleErr}
		provider := NewLibnutcoreKeyboardProvider(client)

		if err := provider.PressKey(context.Background(), shared.KeyA); !errors.Is(err, toggleErr) {
			t.Fatalf("expected press toggle error, got %v", err)
		}
	})

	t.Run("release", func(t *testing.T) {
		toggleErr := errors.New("toggle failed")
		client := &fakeLibnutcoreClient{keyToggleErr: toggleErr}
		provider := NewLibnutcoreKeyboardProvider(client)

		if err := provider.ReleaseKey(context.Background(), shared.KeyA); !errors.Is(err, toggleErr) {
			t.Fatalf("expected release toggle error, got %v", err)
		}
	})
}

func TestMouseProviderScrollAndButtonMapping(t *testing.T) {
	client := &fakeLibnutcoreClient{mousePosition: common.Point{X: 7, Y: 9}}
	provider := NewLibnutcoreMouseProvider(client)

	if err := provider.Click(context.Background(), shared.ButtonLeft); err != nil {
		t.Fatalf("unexpected left click error: %v", err)
	}
	if client.mouseClickButton != common.MouseButtonLeft || client.mouseClickDouble {
		t.Fatalf("unexpected left mouse click call: button=%s double=%v", client.mouseClickButton, client.mouseClickDouble)
	}
	if err := provider.Click(context.Background(), shared.ButtonRight); err != nil {
		t.Fatalf("unexpected click error: %v", err)
	}
	if client.mouseClickButton != common.MouseButtonRight || client.mouseClickDouble {
		t.Fatalf("unexpected mouse click call: button=%s double=%v", client.mouseClickButton, client.mouseClickDouble)
	}
	if err := provider.DoubleClick(context.Background(), shared.ButtonMiddle); err != nil {
		t.Fatalf("unexpected double click error: %v", err)
	}
	if client.mouseClickButton != common.MouseButtonMiddle || client.mouseClickDouble {
		t.Fatalf("unexpected double click call: button=%s double=%v", client.mouseClickButton, client.mouseClickDouble)
	}
	if len(client.mouseClicks) < 4 {
		t.Fatalf("expected double click to emit two ordinary clicks, got %#v", client.mouseClicks)
	}
	if client.mouseClicks[len(client.mouseClicks)-2] != common.MouseButtonMiddle || client.mouseClicks[len(client.mouseClicks)-1] != common.MouseButtonMiddle {
		t.Fatalf("expected final two clicks to target middle button, got %#v", client.mouseClicks)
	}
	if err := provider.ScrollUp(context.Background(), 3); err != nil {
		t.Fatalf("unexpected scroll up error: %v", err)
	}
	if err := provider.ScrollDown(context.Background(), 4); err != nil {
		t.Fatalf("unexpected scroll down error: %v", err)
	}
	if err := provider.ScrollLeft(context.Background(), 5); err != nil {
		t.Fatalf("unexpected scroll left error: %v", err)
	}
	if err := provider.ScrollRight(context.Background(), 6); err != nil {
		t.Fatalf("unexpected scroll right error: %v", err)
	}
	wantScrolls := [][2]int{{0, 3}, {0, -4}, {-5, 0}, {6, 0}}
	if !reflect.DeepEqual(client.scrolls, wantScrolls) {
		t.Fatalf("unexpected scroll mapping: %#v", client.scrolls)
	}
}

func TestMouseProviderPositionAndButtonToggleMapping(t *testing.T) {
	client := &fakeLibnutcoreClient{mousePosition: common.Point{X: 7, Y: 9}}
	provider := NewLibnutcoreMouseProvider(client)

	position, err := provider.CurrentMousePosition(context.Background())
	if err != nil {
		t.Fatalf("unexpected current position error: %v", err)
	}
	if position != (shared.Point{X: 7, Y: 9}) {
		t.Fatalf("unexpected current position: %#v", position)
	}
	if err := provider.SetMousePosition(context.Background(), shared.Point{X: 3, Y: 4}); err != nil {
		t.Fatalf("unexpected set mouse position error: %v", err)
	}
	if client.moveMousePoint != (common.Point{X: 3, Y: 4}) {
		t.Fatalf("unexpected native move target: %#v", client.moveMousePoint)
	}
	if err := provider.PressButton(context.Background(), shared.ButtonRight); err != nil {
		t.Fatalf("unexpected press button error: %v", err)
	}
	if err := provider.ReleaseButton(context.Background(), shared.ButtonRight); err != nil {
		t.Fatalf("unexpected release button error: %v", err)
	}
	if !reflect.DeepEqual(client.mouseToggleStates, []common.ButtonState{common.ButtonStateDown, common.ButtonStateUp}) {
		t.Fatalf("unexpected toggle states: %#v", client.mouseToggleStates)
	}
	if !reflect.DeepEqual(client.mouseToggleButton, []common.MouseButton{common.MouseButtonRight, common.MouseButtonRight}) {
		t.Fatalf("unexpected toggle buttons: %#v", client.mouseToggleButton)
	}
}

func TestMouseProviderRejectsUnknownButtonsAndChecksContext(t *testing.T) {
	client := &fakeLibnutcoreClient{}
	provider := NewLibnutcoreMouseProvider(client)

	if err := provider.Click(context.Background(), shared.Button(-1)); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid click button error, got %v", err)
	}
	if err := provider.PressButton(context.Background(), shared.Button(-1)); !errors.Is(err, common.ErrInvalidToken) {
		t.Fatalf("expected invalid press button error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := provider.ScrollUp(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if len(client.scrolls) != 0 {
		t.Fatalf("expected no scroll calls after cancellation, got %#v", client.scrolls)
	}
}

func TestMouseProviderPropagatesNativeErrors(t *testing.T) {
	t.Run("move", func(t *testing.T) {
		moveErr := errors.New("move failed")
		client := &fakeLibnutcoreClient{moveMouseErr: moveErr}
		provider := NewLibnutcoreMouseProvider(client)

		if err := provider.SetMousePosition(context.Background(), shared.Point{X: 1, Y: 2}); !errors.Is(err, moveErr) {
			t.Fatalf("expected move error, got %v", err)
		}
	})

	t.Run("position", func(t *testing.T) {
		positionErr := errors.New("position failed")
		client := &fakeLibnutcoreClient{mousePositionErr: positionErr}
		provider := NewLibnutcoreMouseProvider(client)

		if _, err := provider.CurrentMousePosition(context.Background()); !errors.Is(err, positionErr) {
			t.Fatalf("expected position error, got %v", err)
		}
	})

	t.Run("click", func(t *testing.T) {
		clickErr := errors.New("click failed")
		client := &fakeLibnutcoreClient{mouseClickErr: clickErr}
		provider := NewLibnutcoreMouseProvider(client)

		if err := provider.Click(context.Background(), shared.ButtonLeft); !errors.Is(err, clickErr) {
			t.Fatalf("expected click error, got %v", err)
		}
	})

	t.Run("toggle", func(t *testing.T) {
		toggleErr := errors.New("toggle failed")
		client := &fakeLibnutcoreClient{mouseToggleErr: toggleErr}
		provider := NewLibnutcoreMouseProvider(client)

		if err := provider.ReleaseButton(context.Background(), shared.ButtonLeft); !errors.Is(err, toggleErr) {
			t.Fatalf("expected toggle error, got %v", err)
		}
	})

	t.Run("scroll", func(t *testing.T) {
		scrollErr := errors.New("scroll failed")
		client := &fakeLibnutcoreClient{scrollErr: scrollErr}
		provider := NewLibnutcoreMouseProvider(client)

		if err := provider.ScrollRight(context.Background(), 1); !errors.Is(err, scrollErr) {
			t.Fatalf("expected scroll error, got %v", err)
		}
	})
}

func TestScreenProviderCaptureConvertsBitmapToSharedImage(t *testing.T) {
	client := &fakeLibnutcoreClient{
		captureBitmap: &common.Bitmap{Width: 20, Height: 10, ByteWidth: 80, BitsPerPixel: 32, BytesPerPixel: 4, Image: []byte{1, 2, 3, 4}},
		screenSize:    common.Size{Width: 10, Height: 5},
	}
	provider := NewLibnutcoreScreenProvider(client)

	image, err := provider.GrabScreen(context.Background())
	if err != nil {
		t.Fatalf("unexpected grab screen error: %v", err)
	}
	if image.ColorMode != shared.ColorModeBGR {
		t.Fatalf("unexpected color mode: %s", image.ColorMode)
	}
	if image.BitsPerPixel != 32 || image.ByteWidth != 80 || image.Channels != 4 {
		t.Fatalf("unexpected bitmap metadata: %#v", image)
	}
	if client.captureRegion != nil {
		t.Fatalf("expected full screen capture to use nil native region, got %#v", client.captureRegion)
	}
	density := image.NormalizedPixelDensity()
	if density.ScaleX != 2 || density.ScaleY != 2 {
		t.Fatalf("unexpected pixel density: %#v", density)
	}
}

func TestScreenProviderRegionCaptureAndHighlightForwarding(t *testing.T) {
	client := &fakeLibnutcoreClient{
		captureBitmap: &common.Bitmap{Width: 12, Height: 8, ByteWidth: 48, BitsPerPixel: 32, BytesPerPixel: 4, Image: []byte{1, 2, 3, 4}},
		screenSize:    common.Size{Width: 20, Height: 10},
	}
	provider := NewLibnutcoreScreenProvider(client)
	region := shared.Region{Left: 3, Top: 4, Width: 4, Height: 2}

	image, err := provider.GrabScreenRegion(context.Background(), region)
	if err != nil {
		t.Fatalf("unexpected grab screen region error: %v", err)
	}
	if client.captureRegion == nil || *client.captureRegion != (common.Rect{X: 3, Y: 4, Width: 4, Height: 2}) {
		t.Fatalf("unexpected capture region forwarding: %#v", client.captureRegion)
	}
	density := image.NormalizedPixelDensity()
	if density.ScaleX != 3 || density.ScaleY != 4 {
		t.Fatalf("unexpected region pixel density: %#v", density)
	}

	if err := provider.HighlightScreenRegion(context.Background(), region, 25*time.Millisecond, 0.75); err != nil {
		t.Fatalf("unexpected highlight error: %v", err)
	}
	if client.highlightRegion != (common.Rect{X: 3, Y: 4, Width: 4, Height: 2}) || client.highlightDuration != 25*time.Millisecond || client.highlightOpacity != 0.75 {
		t.Fatalf("unexpected highlight forwarding: region=%#v duration=%v opacity=%v", client.highlightRegion, client.highlightDuration, client.highlightOpacity)
	}

	width, err := provider.ScreenWidth(context.Background())
	if err != nil || width != 20 {
		t.Fatalf("unexpected screen width result: width=%d err=%v", width, err)
	}
	height, err := provider.ScreenHeight(context.Background())
	if err != nil || height != 10 {
		t.Fatalf("unexpected screen height result: height=%d err=%v", height, err)
	}
	size, err := provider.ScreenSize(context.Background())
	if err != nil || size != (shared.Region{Left: 0, Top: 0, Width: 20, Height: 10}) {
		t.Fatalf("unexpected screen size result: size=%#v err=%v", size, err)
	}
}

func TestScreenProviderFallsBackToMacOSScreencaptureWhenNativeCaptureUnsupported(t *testing.T) {
	t.Run("full screen", func(t *testing.T) {
		client := &fakeLibnutcoreClient{
			info: libnutcore.BackendInfo{Name: libnutcore.BackendName, Platform: "darwin", BindingState: libnutcore.BindingStateLinked},
			capabilities: common.NewCapabilitySet(
				common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityUnsupported, Reason: "native macOS capture is unsupported"},
			),
			screenSize: common.Size{Width: 5, Height: 3},
		}
		var fallbackRegion *common.Rect
		withScreenFallbackTestHooks(t, "darwin", func(ctx context.Context, region *common.Rect) ([]byte, error) {
			if region != nil {
				copy := *region
				fallbackRegion = &copy
			}
			return testPNGBytes(10, 6), nil
		})
		provider := NewLibnutcoreScreenProvider(client)

		image, err := provider.GrabScreen(context.Background())
		if err != nil {
			t.Fatalf("unexpected grab screen fallback error: %v", err)
		}
		if client.captureCalls != 0 {
			t.Fatalf("expected native capture to be skipped, got %d calls", client.captureCalls)
		}
		if fallbackRegion != nil {
			t.Fatalf("expected full screen fallback to omit a region, got %#v", fallbackRegion)
		}
		if image.Width != 10 || image.Height != 6 || image.Channels != 4 || image.ColorMode != shared.ColorModeRGB {
			t.Fatalf("unexpected fallback image metadata: %#v", image)
		}
		density := image.NormalizedPixelDensity()
		if density.ScaleX != 2 || density.ScaleY != 2 {
			t.Fatalf("unexpected full screen fallback density: %#v", density)
		}
	})

	t.Run("region", func(t *testing.T) {
		client := &fakeLibnutcoreClient{
			info: libnutcore.BackendInfo{Name: libnutcore.BackendName, Platform: "darwin", BindingState: libnutcore.BindingStateLinked},
			capabilities: common.NewCapabilitySet(
				common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityUnsupported, Reason: "native macOS capture is unsupported"},
			),
		}
		region := shared.Region{Left: 3, Top: 4, Width: 4, Height: 2}
		var fallbackRegion *common.Rect
		withScreenFallbackTestHooks(t, "darwin", func(ctx context.Context, nativeRegion *common.Rect) ([]byte, error) {
			if nativeRegion == nil {
				t.Fatal("expected region capture to pass a native region to the fallback")
			}
			copy := *nativeRegion
			fallbackRegion = &copy
			return testPNGBytes(8, 4), nil
		})
		provider := NewLibnutcoreScreenProvider(client)

		image, err := provider.GrabScreenRegion(context.Background(), region)
		if err != nil {
			t.Fatalf("unexpected grab screen region fallback error: %v", err)
		}
		if client.captureCalls != 0 {
			t.Fatalf("expected native region capture to be skipped, got %d calls", client.captureCalls)
		}
		if fallbackRegion == nil || *fallbackRegion != (common.Rect{X: 3, Y: 4, Width: 4, Height: 2}) {
			t.Fatalf("unexpected fallback region forwarding: %#v", fallbackRegion)
		}
		density := image.NormalizedPixelDensity()
		if density.ScaleX != 2 || density.ScaleY != 2 {
			t.Fatalf("unexpected region fallback density: %#v", density)
		}
	})
}

func TestScreenProviderReturnsCapabilityUnavailableWhenCaptureUnsupportedOffDarwin(t *testing.T) {
	client := &fakeLibnutcoreClient{
		info: libnutcore.BackendInfo{Name: libnutcore.BackendName, Platform: "linux", BindingState: libnutcore.BindingStateLinked, Notes: []string{"x11 capture path missing"}},
		capabilities: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityUnsupported, Reason: "native capture disabled"},
		),
	}
	fallbackCalled := false
	withScreenFallbackTestHooks(t, "linux", func(ctx context.Context, region *common.Rect) ([]byte, error) {
		fallbackCalled = true
		return nil, errors.New("unexpected fallback invocation")
	})
	provider := NewLibnutcoreScreenProvider(client)

	_, err := provider.GrabScreenRegion(context.Background(), shared.Region{Left: 1, Top: 2, Width: 3, Height: 4})
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "native capture disabled") {
		t.Fatalf("expected error to include the native reason, got %v", err)
	}
	if !strings.Contains(err.Error(), "backend=libnut-core") {
		t.Fatalf("expected error to include backend info, got %v", err)
	}
	if fallbackCalled {
		t.Fatal("expected non-darwin capture to avoid the darwin fallback")
	}
	if client.captureCalls != 0 {
		t.Fatalf("expected native capture to be skipped when unsupported, got %d calls", client.captureCalls)
	}
}

func TestScreenProviderChecksHighlightCapabilityBeforeCallingNative(t *testing.T) {
	client := &fakeLibnutcoreClient{
		info: libnutcore.BackendInfo{Name: libnutcore.BackendName, Platform: "linux", BindingState: libnutcore.BindingStateLinked},
		capabilities: common.NewCapabilitySet(
			common.CapabilityStatus{Capability: common.CapabilityScreenCapture, Availability: common.AvailabilityAvailable},
			common.CapabilityStatus{Capability: common.CapabilityScreenHighlight, Availability: common.AvailabilityUnavailable, Reason: "highlight overlay unavailable"},
		),
	}
	provider := NewLibnutcoreScreenProvider(client)

	err := provider.HighlightScreenRegion(context.Background(), shared.Region{Left: 3, Top: 4, Width: 5, Height: 6}, 25*time.Millisecond, 0.5)
	if !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "highlight overlay unavailable") {
		t.Fatalf("expected error to include the native reason, got %v", err)
	}
	if client.highlightCalls != 0 {
		t.Fatalf("expected native highlight to be skipped, got %d calls", client.highlightCalls)
	}
}

func TestScreenProviderRejectsNilBitmapAndPropagatesErrors(t *testing.T) {
	t.Run("nil bitmap", func(t *testing.T) {
		provider := NewLibnutcoreScreenProvider(&fakeLibnutcoreClient{screenSize: common.Size{Width: 10, Height: 5}})

		if _, err := provider.GrabScreen(context.Background()); err == nil || err.Error() != "bitmap is nil" {
			t.Fatalf("expected nil bitmap error, got %v", err)
		}
	})

	t.Run("capture", func(t *testing.T) {
		captureErr := errors.New("capture failed")
		provider := NewLibnutcoreScreenProvider(&fakeLibnutcoreClient{captureErr: captureErr})

		if _, err := provider.GrabScreen(context.Background()); !errors.Is(err, captureErr) {
			t.Fatalf("expected capture error, got %v", err)
		}
	})

	t.Run("screen size", func(t *testing.T) {
		sizeErr := errors.New("size failed")
		provider := NewLibnutcoreScreenProvider(&fakeLibnutcoreClient{
			captureBitmap: &common.Bitmap{Width: 10, Height: 5, ByteWidth: 40, BitsPerPixel: 32, BytesPerPixel: 4, Image: []byte{1, 2, 3, 4}},
			screenSizeErr: sizeErr,
		})

		if _, err := provider.GrabScreen(context.Background()); !errors.Is(err, sizeErr) {
			t.Fatalf("expected screen size error, got %v", err)
		}
	})

	t.Run("context", func(t *testing.T) {
		provider := NewLibnutcoreScreenProvider(&fakeLibnutcoreClient{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := provider.GrabScreenRegion(ctx, shared.Region{Left: 0, Top: 0, Width: 2, Height: 2}); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation, got %v", err)
		}
	})
}

func TestWindowProviderForwardsWindowOperations(t *testing.T) {
	client := &fakeLibnutcoreClient{
		windows:          []common.WindowHandle{3, 4},
		activeWindow:     4,
		windowTitle:      "Editor",
		windowRect:       common.Rect{X: -2, Y: 5, Width: 10, Height: 8},
		focusResult:      true,
		moveWindowResult: true,
		resizeResult:     true,
	}
	provider := NewLibnutcoreWindowProvider(client)

	windows, err := provider.GetWindows(context.Background())
	if err != nil {
		t.Fatalf("unexpected get windows error: %v", err)
	}
	if !reflect.DeepEqual(windows, []shared.WindowHandle{3, 4}) {
		t.Fatalf("unexpected windows: %#v", windows)
	}

	active, err := provider.GetActiveWindow(context.Background())
	if err != nil || active != 4 {
		t.Fatalf("unexpected active window result: active=%d err=%v", active, err)
	}

	title, err := provider.GetWindowTitle(context.Background(), 9)
	if err != nil || title != "Editor" || client.windowTitleHandle != 9 {
		t.Fatalf("unexpected window title forwarding: title=%q handle=%d err=%v", title, client.windowTitleHandle, err)
	}

	region, err := provider.GetWindowRegion(context.Background(), 9)
	if err != nil || region != (shared.Region{Left: -2, Top: 5, Width: 10, Height: 8}) || client.windowRectHandle != 9 {
		t.Fatalf("unexpected window region forwarding: region=%#v handle=%d err=%v", region, client.windowRectHandle, err)
	}

	focused, err := provider.FocusWindow(context.Background(), 9)
	if err != nil || !focused || client.focusHandle != 9 {
		t.Fatalf("unexpected focus forwarding: focused=%v handle=%d err=%v", focused, client.focusHandle, err)
	}

	moved, err := provider.MoveWindow(context.Background(), 9, shared.Point{X: 11, Y: 12})
	if err != nil || !moved || client.moveWindowHandle != 9 || client.moveWindowOrigin != (common.Point{X: 11, Y: 12}) {
		t.Fatalf("unexpected move forwarding: moved=%v handle=%d origin=%#v err=%v", moved, client.moveWindowHandle, client.moveWindowOrigin, err)
	}

	resized, err := provider.ResizeWindow(context.Background(), 9, shared.Size{Width: 20, Height: 21})
	if err != nil || !resized || client.resizeHandle != 9 || client.resizeSize != (common.Size{Width: 20, Height: 21}) {
		t.Fatalf("unexpected resize forwarding: resized=%v handle=%d size=%#v err=%v", resized, client.resizeHandle, client.resizeSize, err)
	}
}

func TestWindowProviderChecksContextAndPropagatesErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{})
	if _, err := provider.FocusWindow(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	t.Run("get windows", func(t *testing.T) {
		windowsErr := errors.New("windows failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{getWindowsErr: windowsErr})

		if _, err := provider.GetWindows(context.Background()); !errors.Is(err, windowsErr) {
			t.Fatalf("expected get windows error, got %v", err)
		}
	})

	t.Run("active window", func(t *testing.T) {
		activeErr := errors.New("active failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{activeWindowErr: activeErr})

		if _, err := provider.GetActiveWindow(context.Background()); !errors.Is(err, activeErr) {
			t.Fatalf("expected active window error, got %v", err)
		}
	})

	t.Run("title", func(t *testing.T) {
		titleErr := errors.New("title failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{windowTitleErr: titleErr})

		if _, err := provider.GetWindowTitle(context.Background(), 1); !errors.Is(err, titleErr) {
			t.Fatalf("expected title error, got %v", err)
		}
	})

	t.Run("region", func(t *testing.T) {
		regionErr := errors.New("region failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{windowRectErr: regionErr})

		if _, err := provider.GetWindowRegion(context.Background(), 1); !errors.Is(err, regionErr) {
			t.Fatalf("expected region error, got %v", err)
		}
	})

	t.Run("focus", func(t *testing.T) {
		focusErr := errors.New("focus failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{focusErr: focusErr})

		if _, err := provider.FocusWindow(context.Background(), 1); !errors.Is(err, focusErr) {
			t.Fatalf("expected focus error, got %v", err)
		}
	})

	t.Run("move", func(t *testing.T) {
		moveErr := errors.New("move failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{moveWindowErr: moveErr})

		if _, err := provider.MoveWindow(context.Background(), 1, shared.Point{}); !errors.Is(err, moveErr) {
			t.Fatalf("expected move error, got %v", err)
		}
	})

	t.Run("resize", func(t *testing.T) {
		resizeErr := errors.New("resize failed")
		provider := NewLibnutcoreWindowProvider(&fakeLibnutcoreClient{resizeErr: resizeErr})

		if _, err := provider.ResizeWindow(context.Background(), 1, shared.Size{}); !errors.Is(err, resizeErr) {
			t.Fatalf("expected resize error, got %v", err)
		}
	})
}

func TestWindowProviderMinimizeRestoreRemainDeterministic(t *testing.T) {
	client := &fakeLibnutcoreClient{
		minimizeErr: common.CapabilityUnavailable("minimizeWindow", "linux", common.CapabilityWindowMinimize, "libnut-core does not expose a native window minimize primitive"),
		restoreErr:  common.CapabilityUnavailable("restoreWindow", "linux", common.CapabilityWindowRestore, "libnut-core does not expose a native window restore primitive"),
	}
	provider := NewLibnutcoreWindowProvider(client)

	if _, err := provider.MinimizeWindow(context.Background(), 1); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable from minimize, got %v", err)
	}
	if client.minimizeHandle != 1 {
		t.Fatalf("expected minimize handle to be forwarded, got %d", client.minimizeHandle)
	}
	if _, err := provider.RestoreWindow(context.Background(), 1); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected ErrCapabilityUnavailable from restore, got %v", err)
	}
	if client.restoreHandle != 1 {
		t.Fatalf("expected restore handle to be forwarded, got %d", client.restoreHandle)
	}
}

func TestRegisterLibnutcoreProvidersSmoke(t *testing.T) {
	registry := NewRegistry()
	client := &fakeLibnutcoreClient{unavailableMode: true}
	RegisterLibnutcoreProviders(registry, client)

	keyboard, err := registry.Keyboard()
	if err != nil {
		t.Fatalf("unexpected keyboard lookup error: %v", err)
	}
	mouse, err := registry.Mouse()
	if err != nil {
		t.Fatalf("unexpected mouse lookup error: %v", err)
	}
	screen, err := registry.Screen()
	if err != nil {
		t.Fatalf("unexpected screen lookup error: %v", err)
	}
	window, err := registry.Window()
	if err != nil {
		t.Fatalf("unexpected window lookup error: %v", err)
	}
	accessibility, err := registry.Accessibility()
	if err != nil {
		t.Fatalf("unexpected accessibility lookup error: %v", err)
	}

	if err := keyboard.Click(context.Background(), shared.KeyA); err == nil {
		t.Fatal("expected default build keyboard click to report native unavailability")
	}
	if err := mouse.ScrollUp(context.Background(), 1); err == nil {
		t.Fatal("expected default build mouse scroll to report native unavailability")
	}
	if _, err := screen.ScreenSize(context.Background()); err == nil {
		t.Fatal("expected default build screen size to report native unavailability")
	}
	if _, err := window.MinimizeWindow(context.Background(), 1); !errors.Is(err, common.ErrCapabilityUnavailable) {
		t.Fatalf("expected minimize to remain deterministic, got %v", err)
	}
	if !reflect.DeepEqual(accessibility.Capabilities(), client.Capabilities()) {
		t.Fatalf("unexpected accessibility capabilities: %#v", accessibility.Capabilities())
	}
}
