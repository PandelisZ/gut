package provider

import (
	"context"
	"errors"
	"reflect"
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
	info              libnutcore.BackendInfo
	capabilities      common.CapabilitySet
	keyTapKey         string
	keyTapModifiers   []string
	keyTapErr         error
	typedText         []string
	typeStringErr     error
	keyToggles        []recordedToggle
	keyToggleErr      error
	mouseClickButton  common.MouseButton
	mouseClickDouble  bool
	mouseClickErr     error
	mouseToggleStates []common.ButtonState
	mouseToggleButton []common.MouseButton
	mouseToggleErr    error
	scrolls           [][2]int
	scrollErr         error
	moveMousePoint    common.Point
	moveMouseErr      error
	mousePosition     common.Point
	mousePositionErr  error
	captureRegion     *common.Rect
	captureBitmap     *common.Bitmap
	captureErr        error
	screenSize        common.Size
	screenSizeErr     error
	highlightRegion   common.Rect
	highlightDuration time.Duration
	highlightOpacity  float64
	highlightErr      error
	windows           []common.WindowHandle
	getWindowsErr     error
	activeWindow      common.WindowHandle
	activeWindowErr   error
	windowRectHandle  common.WindowHandle
	windowRect        common.Rect
	windowRectErr     error
	windowTitleHandle common.WindowHandle
	windowTitle       string
	windowTitleErr    error
	focusHandle       common.WindowHandle
	focusResult       bool
	focusErr          error
	resizeHandle      common.WindowHandle
	resizeSize        common.Size
	resizeResult      bool
	resizeErr         error
	moveWindowHandle  common.WindowHandle
	moveWindowOrigin  common.Point
	moveWindowResult  bool
	moveWindowErr     error
	unavailableMode   bool
	minimizeHandle    common.WindowHandle
	minimizeErr       error
	restoreHandle     common.WindowHandle
	restoreErr        error
}

func (f *fakeLibnutcoreClient) Info() libnutcore.BackendInfo { return f.info }
func (f *fakeLibnutcoreClient) Capabilities() common.CapabilitySet {
	return f.capabilities
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
	f.highlightRegion = region
	f.highlightDuration = duration
	f.highlightOpacity = opacity
	return f.highlightErr
}
func (f *fakeLibnutcoreClient) CaptureScreen(region *common.Rect) (*common.Bitmap, error) {
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
	if client.keyTapKey != "a" {
		t.Fatalf("unexpected primary key: %q", client.keyTapKey)
	}
	if !reflect.DeepEqual(client.keyTapModifiers, []string{"control", "right_alt"}) {
		t.Fatalf("unexpected modifiers: %#v", client.keyTapModifiers)
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
	if client.keyTapKey != "" {
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
	if client.keyTapKey != "delete" {
		t.Fatalf("unexpected delete token: %q", client.keyTapKey)
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
		tapErr := errors.New("tap failed")
		client := &fakeLibnutcoreClient{keyTapErr: tapErr}
		provider := NewLibnutcoreKeyboardProvider(client)

		if err := provider.Click(context.Background(), shared.KeyLeftShift, shared.KeyA); !errors.Is(err, tapErr) {
			t.Fatalf("expected tap error, got %v", err)
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
	if client.mouseClickButton != common.MouseButtonMiddle || !client.mouseClickDouble {
		t.Fatalf("unexpected double click call: button=%s double=%v", client.mouseClickButton, client.mouseClickDouble)
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
}
