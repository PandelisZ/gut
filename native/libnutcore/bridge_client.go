//go:build cgo

package libnutcore

import (
	"time"

	"github.com/PandelisZ/gut/native/common"
)

var _ Client = (*bridgeClient)(nil)

type bridgeClient struct {
	platform      string
	bindingState  BindingState
	notes         []string
	capabilities  common.CapabilitySet
	mouseDelay    time.Duration
	keyboardDelay time.Duration
}

func newBridgeClient(platform string, notes []string, capabilities common.CapabilitySet, options Options) Client {
	client := &bridgeClient{
		platform:      platform,
		bindingState:  BindingStateLinked,
		notes:         append([]string(nil), notes...),
		capabilities:  cloneCapabilities(capabilities),
		mouseDelay:    options.MouseDelay,
		keyboardDelay: options.KeyboardDelay,
	}
	if platform == "linux" && options.XDisplayName != "" {
		_ = client.SetXDisplayName(options.XDisplayName)
	}
	return client
}

func (c *bridgeClient) Info() BackendInfo {
	return BackendInfo{
		Name:         BackendName,
		Platform:     c.platform,
		BindingState: c.bindingState,
		Notes:        append([]string(nil), c.notes...),
	}
}

func (c *bridgeClient) Capabilities() common.CapabilitySet {
	return cloneCapabilities(c.capabilities)
}

func (c *bridgeClient) GetPermissionSnapshot() (common.PermissionSnapshot, error) {
	return bridgeGetPermissionSnapshot()
}

func (c *bridgeClient) GetFocusedWindow() (common.FocusedWindowMetadata, error) {
	return bridgeGetFocusedWindow()
}

func (c *bridgeClient) RaiseFocusedWindow() error {
	return bridgeRaiseFocusedWindow()
}

func (c *bridgeClient) GetFocusedElement() (common.UIElementMetadata, error) {
	return bridgeGetFocusedElement()
}

func (c *bridgeClient) PerformFocusedElementAction(action common.AXAction) error {
	return bridgePerformFocusedElementAction(action)
}

func (c *bridgeClient) GetElementAtPoint(position common.Point) (common.UIElementMetadata, error) {
	return bridgeGetElementAtPoint(position)
}

func (c *bridgeClient) PerformElementActionAtPoint(position common.Point, action common.AXAction) error {
	return bridgePerformElementActionAtPoint(position, action)
}

func (c *bridgeClient) FocusElementAtPoint(position common.Point) error {
	return bridgeFocusElementAtPoint(position)
}

func (c *bridgeClient) SearchAXElements(query common.AXElementSearchQuery) ([]common.AXElementMatch, error) {
	return bridgeSearchAXElements(query)
}

func (c *bridgeClient) FocusAXElement(ref common.AXElementRef) error {
	return bridgeFocusAXElement(ref)
}

func (c *bridgeClient) PerformAXElementAction(ref common.AXElementRef, action common.AXAction) error {
	return bridgePerformAXElementAction(ref, action)
}

func (c *bridgeClient) DragMouse(position common.Point, button common.MouseButton) error {
	return bridgeDragMouse(position, button)
}

func (c *bridgeClient) MoveMouse(position common.Point) error {
	return bridgeMoveMouse(position)
}

func (c *bridgeClient) GetMousePosition() (common.Point, error) {
	return bridgeGetMousePosition()
}

func (c *bridgeClient) MouseClick(button common.MouseButton, double bool) error {
	return bridgeMouseClick(button, double)
}

func (c *bridgeClient) MouseToggle(state common.ButtonState, button common.MouseButton) error {
	return bridgeMouseToggle(state, button)
}

func (c *bridgeClient) ScrollMouse(horizontal, vertical int) error {
	return bridgeScrollMouse(horizontal, vertical)
}

func (c *bridgeClient) SetMouseDelay(delay time.Duration) error {
	c.mouseDelay = delay
	return bridgeSetMouseDelay(delay)
}

func (c *bridgeClient) KeyTap(key string, modifiers ...string) error {
	return bridgeKeyTap(key, modifiers)
}

func (c *bridgeClient) KeyToggle(key string, state common.KeyState, modifiers ...string) error {
	return bridgeKeyToggle(key, state, modifiers)
}

func (c *bridgeClient) TypeString(text string) error {
	return bridgeTypeString(text)
}

func (c *bridgeClient) TypeStringDelayed(text string, cpm int) error {
	return bridgeTypeStringDelayed(text, cpm)
}

func (c *bridgeClient) SetKeyboardDelay(delay time.Duration) error {
	c.keyboardDelay = delay
	return bridgeSetKeyboardDelay(delay)
}

func (c *bridgeClient) GetScreenSize() (common.Size, error) {
	return bridgeGetScreenSize()
}

func (c *bridgeClient) Highlight(region common.Rect, duration time.Duration, opacity float64) error {
	return bridgeHighlight(region, duration, opacity)
}

func (c *bridgeClient) CaptureScreen(region *common.Rect) (*common.Bitmap, error) {
	status := c.capabilities.Status(common.CapabilityScreenCapture)
	if status.Availability != common.AvailabilityAvailable {
		return nil, common.CapabilityUnavailable("captureScreen", c.platform, common.CapabilityScreenCapture, status.Reason)
	}
	return bridgeCaptureScreen(region)
}

func (c *bridgeClient) GetWindows() ([]common.WindowHandle, error) {
	return bridgeGetWindows()
}

func (c *bridgeClient) GetActiveWindow() (common.WindowHandle, error) {
	return bridgeGetActiveWindow()
}

func (c *bridgeClient) GetWindowRect(handle common.WindowHandle) (common.Rect, error) {
	return bridgeGetWindowRect(handle)
}

func (c *bridgeClient) GetWindowTitle(handle common.WindowHandle) (string, error) {
	return bridgeGetWindowTitle(handle)
}

func (c *bridgeClient) FocusWindow(handle common.WindowHandle) (bool, error) {
	return bridgeFocusWindow(handle)
}

func (c *bridgeClient) ResizeWindow(handle common.WindowHandle, size common.Size) (bool, error) {
	return bridgeResizeWindow(handle, size)
}

func (c *bridgeClient) MoveWindow(handle common.WindowHandle, origin common.Point) (bool, error) {
	return bridgeMoveWindow(handle, origin)
}

func (c *bridgeClient) MinimizeWindow(handle common.WindowHandle) (bool, error) {
	status := c.capabilities.Status(common.CapabilityWindowMinimize)
	return false, common.CapabilityUnavailable("minimizeWindow", c.platform, common.CapabilityWindowMinimize, status.Reason)
}

func (c *bridgeClient) RestoreWindow(handle common.WindowHandle) (bool, error) {
	status := c.capabilities.Status(common.CapabilityWindowRestore)
	return false, common.CapabilityUnavailable("restoreWindow", c.platform, common.CapabilityWindowRestore, status.Reason)
}

func (c *bridgeClient) GetXDisplayName() (string, error) {
	status := c.capabilities.Status(common.CapabilityX11DisplayGet)
	if status.Availability == common.AvailabilityUnsupported {
		return "", common.UnsupportedOperation("getXDisplayName", c.platform, common.CapabilityX11DisplayGet, status.Reason)
	}
	return bridgeGetXDisplayName()
}

func (c *bridgeClient) SetXDisplayName(name string) error {
	status := c.capabilities.Status(common.CapabilityX11DisplaySet)
	if status.Availability == common.AvailabilityUnsupported {
		return common.UnsupportedOperation("setXDisplayName", c.platform, common.CapabilityX11DisplaySet, status.Reason)
	}
	return bridgeSetXDisplayName(name)
}
