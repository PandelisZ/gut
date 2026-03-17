package libnutcore

import (
	"time"

	"gut/native/common"
)

var _ Client = (*unavailableClient)(nil)

type unavailableClient struct {
	platform      string
	bindingState  BindingState
	notes         []string
	capabilities  common.CapabilitySet
	mouseDelay    time.Duration
	keyboardDelay time.Duration
	xDisplayName  string
}

func newUnavailableClient(platform string, notes []string, capabilities common.CapabilitySet, options Options) Client {
	return &unavailableClient{
		platform:      platform,
		bindingState:  BindingStateUnavailable,
		notes:         append([]string(nil), notes...),
		capabilities:  cloneCapabilities(capabilities),
		mouseDelay:    options.MouseDelay,
		keyboardDelay: options.KeyboardDelay,
		xDisplayName:  options.XDisplayName,
	}
}

func (c *unavailableClient) Info() BackendInfo {
	return BackendInfo{
		Name:         BackendName,
		Platform:     c.platform,
		BindingState: c.bindingState,
		Notes:        append([]string(nil), c.notes...),
	}
}

func (c *unavailableClient) Capabilities() common.CapabilitySet {
	return cloneCapabilities(c.capabilities)
}

func (c *unavailableClient) GetPermissionSnapshot() (common.PermissionSnapshot, error) {
	return common.PermissionSnapshot{}, c.unavailable("getPermissionSnapshot", common.CapabilityPermissionReadiness)
}

func (c *unavailableClient) GetFocusedWindow() (common.FocusedWindowMetadata, error) {
	return common.FocusedWindowMetadata{}, c.unavailable("getFocusedWindow", common.CapabilityAXFocusedWindowMetadata)
}

func (c *unavailableClient) RaiseFocusedWindow() error {
	return c.capabilityError("raiseFocusedWindow", common.CapabilityAXFocusedWindowRaise)
}

func (c *unavailableClient) GetFocusedElement() (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, c.unavailable("getFocusedElement", common.CapabilityAXFocusedElementMetadata)
}

func (c *unavailableClient) PerformFocusedElementAction(action common.AXAction) error {
	return c.capabilityError("performFocusedElementAction", common.CapabilityAXFocusedElementAction)
}

func (c *unavailableClient) GetElementAtPoint(position common.Point) (common.UIElementMetadata, error) {
	return common.UIElementMetadata{}, c.unavailable("getElementAtPoint", common.CapabilityAXElementAtPointMetadata)
}

func (c *unavailableClient) PerformElementActionAtPoint(position common.Point, action common.AXAction) error {
	return c.capabilityError("performElementActionAtPoint", common.CapabilityAXElementActionAtPoint)
}

func (c *unavailableClient) FocusElementAtPoint(position common.Point) error {
	return c.capabilityError("focusElementAtPoint", common.CapabilityAXElementFocusAtPoint)
}

func (c *unavailableClient) DragMouse(position common.Point, button common.MouseButton) error {
	return c.unavailable("dragMouse", common.CapabilityMouseDrag)
}

func (c *unavailableClient) MoveMouse(position common.Point) error {
	return c.unavailable("moveMouse", common.CapabilityMouseMove)
}

func (c *unavailableClient) GetMousePosition() (common.Point, error) {
	return common.Point{}, c.unavailable("getMousePos", common.CapabilityMousePosition)
}

func (c *unavailableClient) MouseClick(button common.MouseButton, double bool) error {
	return c.unavailable("mouseClick", common.CapabilityMouseClick)
}

func (c *unavailableClient) MouseToggle(state common.ButtonState, button common.MouseButton) error {
	return c.unavailable("mouseToggle", common.CapabilityMouseToggle)
}

func (c *unavailableClient) ScrollMouse(horizontal, vertical int) error {
	return c.unavailable("scrollMouse", common.CapabilityMouseScroll)
}

func (c *unavailableClient) SetMouseDelay(delay time.Duration) error {
	c.mouseDelay = delay
	return nil
}

func (c *unavailableClient) KeyTap(key string, modifiers ...string) error {
	return c.unavailable("keyTap", common.CapabilityKeyboardTap)
}

func (c *unavailableClient) KeyToggle(key string, state common.KeyState, modifiers ...string) error {
	return c.unavailable("keyToggle", common.CapabilityKeyboardToggle)
}

func (c *unavailableClient) TypeString(text string) error {
	return c.unavailable("typeString", common.CapabilityKeyboardType)
}

func (c *unavailableClient) TypeStringDelayed(text string, cpm int) error {
	return c.unavailable("typeStringDelayed", common.CapabilityKeyboardType)
}

func (c *unavailableClient) SetKeyboardDelay(delay time.Duration) error {
	c.keyboardDelay = delay
	return nil
}

func (c *unavailableClient) GetScreenSize() (common.Size, error) {
	return common.Size{}, c.unavailable("getScreenSize", common.CapabilityScreenSize)
}

func (c *unavailableClient) Highlight(region common.Rect, duration time.Duration, opacity float64) error {
	return c.unavailable("highlight", common.CapabilityScreenHighlight)
}

func (c *unavailableClient) CaptureScreen(region *common.Rect) (*common.Bitmap, error) {
	return nil, c.unavailable("captureScreen", common.CapabilityScreenCapture)
}

func (c *unavailableClient) GetWindows() ([]common.WindowHandle, error) {
	return nil, c.unavailable("getWindows", common.CapabilityWindowList)
}

func (c *unavailableClient) GetActiveWindow() (common.WindowHandle, error) {
	return 0, c.unavailable("getActiveWindow", common.CapabilityWindowActive)
}

func (c *unavailableClient) GetWindowRect(handle common.WindowHandle) (common.Rect, error) {
	return common.Rect{}, c.unavailable("getWindowRect", common.CapabilityWindowRect)
}

func (c *unavailableClient) GetWindowTitle(handle common.WindowHandle) (string, error) {
	return "", c.unavailable("getWindowTitle", common.CapabilityWindowTitle)
}

func (c *unavailableClient) FocusWindow(handle common.WindowHandle) (bool, error) {
	return false, c.unavailable("focusWindow", common.CapabilityWindowFocus)
}

func (c *unavailableClient) ResizeWindow(handle common.WindowHandle, size common.Size) (bool, error) {
	return false, c.unavailable("resizeWindow", common.CapabilityWindowResize)
}

func (c *unavailableClient) MoveWindow(handle common.WindowHandle, origin common.Point) (bool, error) {
	return false, c.unavailable("moveWindow", common.CapabilityWindowMove)
}

func (c *unavailableClient) MinimizeWindow(handle common.WindowHandle) (bool, error) {
	return false, c.capabilityError("minimizeWindow", common.CapabilityWindowMinimize)
}

func (c *unavailableClient) RestoreWindow(handle common.WindowHandle) (bool, error) {
	return false, c.capabilityError("restoreWindow", common.CapabilityWindowRestore)
}

func (c *unavailableClient) GetXDisplayName() (string, error) {
	status := c.capabilities.Status(common.CapabilityX11DisplayGet)
	if status.Availability == common.AvailabilityUnsupported {
		return "", common.UnsupportedOperation("getXDisplayName", c.platform, common.CapabilityX11DisplayGet, status.Reason)
	}
	return c.xDisplayName, nil
}

func (c *unavailableClient) SetXDisplayName(name string) error {
	status := c.capabilities.Status(common.CapabilityX11DisplaySet)
	if status.Availability == common.AvailabilityUnsupported {
		return common.UnsupportedOperation("setXDisplayName", c.platform, common.CapabilityX11DisplaySet, status.Reason)
	}
	c.xDisplayName = name
	return nil
}

func (c *unavailableClient) unavailable(operation string, capability common.Capability) error {
	status := c.capabilities.Status(capability)
	detail := status.Reason
	if detail == "" {
		detail = "native backend unavailable"
	}
	if status.Availability == common.AvailabilityPermissionBlocked {
		return common.PermissionDeniedOperation(operation, c.platform, capability, detail)
	}
	if status.Availability == common.AvailabilityUnsupported {
		return common.UnsupportedOperation(operation, c.platform, capability, detail)
	}
	return common.UnavailableOperation(operation, c.platform, capability, detail)
}

func (c *unavailableClient) capabilityError(operation string, capability common.Capability) error {
	status := c.capabilities.Status(capability)
	detail := status.Reason
	if detail == "" {
		detail = "native capability unavailable"
	}
	if status.Availability == common.AvailabilityPermissionBlocked {
		return common.PermissionDeniedOperation(operation, c.platform, capability, detail)
	}
	if status.Availability == common.AvailabilityUnsupported {
		return common.CapabilityUnavailable(operation, c.platform, capability, detail)
	}
	return common.UnavailableOperation(operation, c.platform, capability, detail)
}

func cloneCapabilities(set common.CapabilitySet) common.CapabilitySet {
	clone := make(common.CapabilitySet, len(set))
	for capability, status := range set {
		clone[capability] = status
	}
	return clone
}
