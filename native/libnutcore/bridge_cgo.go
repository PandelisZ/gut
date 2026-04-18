//go:build cgo

package libnutcore

/*
#cgo linux CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo linux CXXFLAGS: -std=c++17 -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo linux LDFLAGS: -lX11 -lXtst
#cgo windows CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo windows CXXFLAGS: -std=c++17 -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo windows LDFLAGS: -lgdi32 -luser32
#cgo darwin CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo darwin CXXFLAGS: -x objective-c++ -std=c++17 -I${SRCDIR} -I${SRCDIR}/../../../libnut-core/src
#cgo darwin LDFLAGS: -framework Cocoa -framework Foundation -framework AppKit -framework ApplicationServices -framework Carbon -framework IOKit
#include <stdlib.h>
#include "bridge_shim.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/PandelisZ/gut/native/common"
)

func bridgeDragMouse(position common.Point, button common.MouseButton) error {
	cButton := C.CString(string(button))
	defer C.free(unsafe.Pointer(cButton))
	return bridgeStatusError("dragMouse", common.CapabilityMouseDrag, int(C.gut_drag_mouse(C.int64_t(position.X), C.int64_t(position.Y), cButton)))
}

func bridgeMoveMouse(position common.Point) error {
	return bridgeStatusError("moveMouse", common.CapabilityMouseMove, int(C.gut_move_mouse(C.int64_t(position.X), C.int64_t(position.Y))))
}

func bridgeGetMousePosition() (common.Point, error) {
	var point C.gut_point
	if err := bridgeStatusError("getMousePos", common.CapabilityMousePosition, int(C.gut_get_mouse_pos(&point))); err != nil {
		return common.Point{}, err
	}
	return common.Point{X: int(point.x), Y: int(point.y)}, nil
}

func bridgeMouseClick(button common.MouseButton, double bool) error {
	cButton := C.CString(string(button))
	defer C.free(unsafe.Pointer(cButton))
	doubleFlag := 0
	if double {
		doubleFlag = 1
	}
	return bridgeStatusError("mouseClick", common.CapabilityMouseClick, int(C.gut_mouse_click(cButton, C.int(doubleFlag))))
}

func bridgeMouseToggle(state common.ButtonState, button common.MouseButton) error {
	cState := C.CString(string(state))
	defer C.free(unsafe.Pointer(cState))
	cButton := C.CString(string(button))
	defer C.free(unsafe.Pointer(cButton))
	return bridgeStatusError("mouseToggle", common.CapabilityMouseToggle, int(C.gut_mouse_toggle(cState, cButton)))
}

func bridgeScrollMouse(horizontal, vertical int) error {
	return bridgeStatusError("scrollMouse", common.CapabilityMouseScroll, int(C.gut_scroll_mouse(C.int(horizontal), C.int(vertical))))
}

func bridgeSetMouseDelay(delay time.Duration) error {
	return bridgeStatusError("setMouseDelay", common.CapabilityMouseDelay, int(C.gut_set_mouse_delay(C.int64_t(delay/time.Millisecond))))
}

func bridgeKeyTap(key string, modifiers []string) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cModifiers, freeModifiers := makeCStringArray(modifiers)
	defer freeModifiers()
	return bridgeStatusError("keyTap", common.CapabilityKeyboardTap, int(C.gut_key_tap(cKey, cModifiers, C.size_t(len(modifiers)))))
}

func bridgeKeyToggle(key string, state common.KeyState, modifiers []string) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cState := C.CString(string(state))
	defer C.free(unsafe.Pointer(cState))
	cModifiers, freeModifiers := makeCStringArray(modifiers)
	defer freeModifiers()
	return bridgeStatusError("keyToggle", common.CapabilityKeyboardToggle, int(C.gut_key_toggle(cKey, cState, cModifiers, C.size_t(len(modifiers)))))
}

func bridgeTypeString(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return bridgeStatusError("typeString", common.CapabilityKeyboardType, int(C.gut_type_string(cText)))
}

func bridgeTypeStringDelayed(text string, cpm int) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return bridgeStatusError("typeStringDelayed", common.CapabilityKeyboardType, int(C.gut_type_string_delayed(cText, C.int(cpm))))
}

func bridgeSetKeyboardDelay(delay time.Duration) error {
	return bridgeStatusError("setKeyboardDelay", common.CapabilityKeyboardDelay, int(C.gut_set_keyboard_delay(C.int64_t(delay/time.Millisecond))))
}

func bridgeGetPermissionSnapshot() (common.PermissionSnapshot, error) {
	var snapshot C.gut_permission_snapshot
	if err := bridgeStatusError("getPermissionSnapshot", common.CapabilityPermissionReadiness, int(C.gut_get_permission_snapshot(&snapshot))); err != nil {
		return common.PermissionSnapshot{}, err
	}
	result := common.PermissionSnapshot{
		Accessibility: common.PermissionStatus{
			Granted:   snapshot.accessibility_granted != 0,
			Supported: snapshot.accessibility_supported != 0,
		},
		ScreenRecording: common.PermissionStatus{
			Granted:   snapshot.screen_recording_granted != 0,
			Supported: snapshot.screen_recording_supported != 0,
		},
	}
	if result.Accessibility.Supported {
		if result.Accessibility.Granted {
			result.Accessibility.Reason = "Accessibility permission is granted"
		} else {
			result.Accessibility.Reason = "Accessibility permission is not granted"
		}
	}
	if result.ScreenRecording.Supported {
		if result.ScreenRecording.Granted {
			result.ScreenRecording.Reason = "Screen Recording permission is granted"
		} else {
			result.ScreenRecording.Reason = "Screen Recording permission is not granted"
		}
	} else {
		result.ScreenRecording.Reason = "Screen Recording preflight is not available in the linked SDK/runtime"
	}
	return result, nil
}

func bridgeGetFocusedWindow() (common.FocusedWindowMetadata, error) {
	var metadata C.gut_window_metadata
	if err := bridgeStatusError("getFocusedWindow", common.CapabilityAXFocusedWindowMetadata, int(C.gut_get_focused_window_metadata(&metadata))); err != nil {
		return common.FocusedWindowMetadata{}, err
	}
	defer C.gut_free_window_metadata(&metadata)
	return common.FocusedWindowMetadata{
		Handle:    common.WindowHandle(metadata.handle),
		Title:     cStringToGo(metadata.title),
		Role:      cStringToGo(metadata.role),
		Subrole:   cStringToGo(metadata.subrole),
		Rect:      common.Rect{X: int(metadata.rect.x), Y: int(metadata.rect.y), Width: int(metadata.rect.width), Height: int(metadata.rect.height)},
		RectKnown: metadata.has_rect != 0,
		Focused:   metadata.focused != 0,
		Main:      metadata.main != 0,
		Minimized: metadata.minimized != 0,
		OwnerPID:  int(metadata.owner_pid),
		OwnerName: cStringToGo(metadata.owner_name),
		BundleID:  cStringToGo(metadata.bundle_id),
	}, nil
}

func bridgeRaiseFocusedWindow() error {
	return bridgeAXInteractionError("raiseFocusedWindow", common.CapabilityAXFocusedWindowRaise, int(C.gut_raise_focused_window()), "no focused AX window is available or the focused window does not support AXRaise")
}

func bridgeGetFocusedElement() (common.UIElementMetadata, error) {
	var metadata C.gut_element_metadata
	if err := bridgeStatusError("getFocusedElement", common.CapabilityAXFocusedElementMetadata, int(C.gut_get_focused_element_metadata(&metadata))); err != nil {
		return common.UIElementMetadata{}, err
	}
	defer C.gut_free_element_metadata(&metadata)
	return convertElementMetadata(metadata), nil
}

func bridgePerformFocusedElementAction(action common.AXAction) error {
	if err := validateAXAction(action, "performFocusedElementAction", common.CapabilityAXFocusedElementAction); err != nil {
		return err
	}
	cAction := C.CString(string(action))
	defer C.free(unsafe.Pointer(cAction))
	return bridgeAXInteractionError("performFocusedElementAction", common.CapabilityAXFocusedElementAction, int(C.gut_perform_focused_element_action(cAction)), "no focused AX element is available or it does not support the requested AX action")
}

func bridgeGetElementAtPoint(position common.Point) (common.UIElementMetadata, error) {
	var metadata C.gut_element_metadata
	if err := bridgeStatusError("getElementAtPoint", common.CapabilityAXElementAtPointMetadata, int(C.gut_get_element_metadata_at_point(C.int64_t(position.X), C.int64_t(position.Y), &metadata))); err != nil {
		return common.UIElementMetadata{}, err
	}
	defer C.gut_free_element_metadata(&metadata)
	return convertElementMetadata(metadata), nil
}

func bridgePerformElementActionAtPoint(position common.Point, action common.AXAction) error {
	if err := validateAXAction(action, "performElementActionAtPoint", common.CapabilityAXElementActionAtPoint); err != nil {
		return err
	}
	cAction := C.CString(string(action))
	defer C.free(unsafe.Pointer(cAction))
	return bridgeAXInteractionError("performElementActionAtPoint", common.CapabilityAXElementActionAtPoint, int(C.gut_perform_element_action_at_point(C.int64_t(position.X), C.int64_t(position.Y), cAction)), "no AX element is available at the requested point or it does not support the requested AX action")
}

func bridgeFocusElementAtPoint(position common.Point) error {
	return bridgeAXInteractionError("focusElementAtPoint", common.CapabilityAXElementFocusAtPoint, int(C.gut_focus_element_at_point(C.int64_t(position.X), C.int64_t(position.Y))), "no AX element is available at the requested point or it does not expose a settable AXFocused attribute")
}

func bridgeSearchAXElements(query common.AXElementSearchQuery) ([]common.AXElementMatch, error) {
	if err := validateAXElementSearchQuery(query); err != nil {
		return nil, err
	}
	cQuery, freeQuery := makeAXElementSearchQuery(query)
	defer freeQuery()
	var matches C.gut_ax_element_match_list
	if err := bridgeStatusError("searchAXElements", common.CapabilityAXElementSearch, int(C.gut_search_ax_elements(&cQuery, &matches))); err != nil {
		return nil, err
	}
	defer C.gut_free_ax_element_match_list(&matches)
	return convertAXElementMatchList(matches), nil
}

func bridgeFocusAXElement(ref common.AXElementRef) error {
	if err := validateAXElementRef(ref, "focusAXElement", common.CapabilityAXElementFocusMatch); err != nil {
		return err
	}
	cRef, freeRef := makeAXElementRef(ref)
	defer freeRef()
	return bridgeAXInteractionError("focusAXElement", common.CapabilityAXElementFocusMatch, int(C.gut_focus_ax_element(&cRef)), "the AX element reference could not be resolved or the element does not expose a settable AXFocused attribute")
}

func bridgePerformAXElementAction(ref common.AXElementRef, action common.AXAction) error {
	if err := validateAXElementRef(ref, "performAXElementAction", common.CapabilityAXElementActionMatch); err != nil {
		return err
	}
	if err := validateAXAction(action, "performAXElementAction", common.CapabilityAXElementActionMatch); err != nil {
		return err
	}
	cRef, freeRef := makeAXElementRef(ref)
	defer freeRef()
	cAction := C.CString(string(action))
	defer C.free(unsafe.Pointer(cAction))
	return bridgeAXInteractionError("performAXElementAction", common.CapabilityAXElementActionMatch, int(C.gut_perform_ax_element_action(&cRef, cAction)), "the AX element reference could not be resolved or the element does not support the requested AX action")
}

func bridgeGetScreenSize() (common.Size, error) {
	var size C.gut_size
	if err := bridgeStatusError("getScreenSize", common.CapabilityScreenSize, int(C.gut_get_screen_size(&size))); err != nil {
		return common.Size{}, err
	}
	return common.Size{Width: int(size.width), Height: int(size.height)}, nil
}

func bridgeHighlight(region common.Rect, duration time.Duration, opacity float64) error {
	return bridgeStatusError("highlight", common.CapabilityScreenHighlight, int(C.gut_highlight(
		C.int64_t(region.X),
		C.int64_t(region.Y),
		C.int64_t(region.Width),
		C.int64_t(region.Height),
		C.int64_t(duration/time.Millisecond),
		C.double(opacity),
	)))
}

func bridgeCaptureScreen(region *common.Rect) (*common.Bitmap, error) {
	var cRegion *C.gut_rect
	if region != nil {
		local := C.gut_rect{x: C.int64_t(region.X), y: C.int64_t(region.Y), width: C.int64_t(region.Width), height: C.int64_t(region.Height)}
		cRegion = &local
	}
	var bitmap *C.gut_bitmap
	if err := bridgeStatusError("captureScreen", common.CapabilityScreenCapture, int(C.gut_capture_screen(cRegion, &bitmap))); err != nil {
		return nil, err
	}
	defer C.gut_free_bitmap(bitmap)
	image := C.GoBytes(unsafe.Pointer(bitmap.image), C.int(bitmap.image_len))
	return &common.Bitmap{
		Width:         int(bitmap.width),
		Height:        int(bitmap.height),
		ByteWidth:     int(bitmap.byte_width),
		BitsPerPixel:  int(bitmap.bits_per_pixel),
		BytesPerPixel: int(bitmap.bytes_per_pixel),
		Image:         image,
	}, nil
}

func bridgeGetWindows() ([]common.WindowHandle, error) {
	var windows C.gut_window_list
	if err := bridgeStatusError("getWindows", common.CapabilityWindowList, int(C.gut_get_windows(&windows))); err != nil {
		return nil, err
	}
	defer C.gut_free_window_list(&windows)
	handles := make([]common.WindowHandle, int(windows.length))
	if windows.length == 0 || windows.handles == nil {
		return handles, nil
	}
	values := unsafe.Slice((*C.int64_t)(windows.handles), int(windows.length))
	for index, handle := range values {
		handles[index] = common.WindowHandle(handle)
	}
	return handles, nil
}

func bridgeGetActiveWindow() (common.WindowHandle, error) {
	var handle C.int64_t
	if err := bridgeStatusError("getActiveWindow", common.CapabilityWindowActive, int(C.gut_get_active_window(&handle))); err != nil {
		return 0, err
	}
	return common.WindowHandle(handle), nil
}

func bridgeGetWindowRect(handle common.WindowHandle) (common.Rect, error) {
	var rect C.gut_rect
	if err := bridgeStatusError("getWindowRect", common.CapabilityWindowRect, int(C.gut_get_window_rect(C.int64_t(handle), &rect))); err != nil {
		return common.Rect{}, err
	}
	return common.Rect{X: int(rect.x), Y: int(rect.y), Width: int(rect.width), Height: int(rect.height)}, nil
}

func bridgeGetWindowTitle(handle common.WindowHandle) (string, error) {
	var title *C.char
	if err := bridgeStatusError("getWindowTitle", common.CapabilityWindowTitle, int(C.gut_get_window_title(C.int64_t(handle), &title))); err != nil {
		return "", err
	}
	defer C.gut_free_string(title)
	return C.GoString(title), nil
}

func bridgeFocusWindow(handle common.WindowHandle) (bool, error) {
	var result C.int
	if err := bridgeStatusError("focusWindow", common.CapabilityWindowFocus, int(C.gut_focus_window(C.int64_t(handle), &result))); err != nil {
		return false, err
	}
	return result != 0, nil
}

func bridgeResizeWindow(handle common.WindowHandle, size common.Size) (bool, error) {
	var result C.int
	if err := bridgeStatusError("resizeWindow", common.CapabilityWindowResize, int(C.gut_resize_window(C.int64_t(handle), C.int64_t(size.Width), C.int64_t(size.Height), &result))); err != nil {
		return false, err
	}
	return result != 0, nil
}

func bridgeMoveWindow(handle common.WindowHandle, origin common.Point) (bool, error) {
	var result C.int
	if err := bridgeStatusError("moveWindow", common.CapabilityWindowMove, int(C.gut_move_window(C.int64_t(handle), C.int64_t(origin.X), C.int64_t(origin.Y), &result))); err != nil {
		return false, err
	}
	return result != 0, nil
}

func bridgeGetXDisplayName() (string, error) {
	var name *C.char
	if err := bridgeStatusError("getXDisplayName", common.CapabilityX11DisplayGet, int(C.gut_get_x_display_name(&name))); err != nil {
		return "", err
	}
	defer C.gut_free_string(name)
	return C.GoString(name), nil
}

func bridgeSetXDisplayName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return bridgeStatusError("setXDisplayName", common.CapabilityX11DisplaySet, int(C.gut_set_x_display_name(cName)))
}

func makeCStringArray(values []string) (**C.char, func()) {
	if len(values) == 0 {
		return nil, func() {}
	}
	cValues := make([]*C.char, len(values))
	for index, value := range values {
		cValues[index] = C.CString(value)
	}
	return (**C.char)(unsafe.Pointer(&cValues[0])), func() {
		for _, value := range cValues {
			C.free(unsafe.Pointer(value))
		}
	}
}

func convertElementMetadata(metadata C.gut_element_metadata) common.UIElementMetadata {
	actions := make([]string, 0, int(metadata.actions.length))
	if metadata.actions.length > 0 && metadata.actions.items != nil {
		values := unsafe.Slice((**C.char)(unsafe.Pointer(metadata.actions.items)), int(metadata.actions.length))
		for _, value := range values {
			actions = append(actions, cStringToGo(value))
		}
	}
	return common.UIElementMetadata{
		Role:        cStringToGo(metadata.role),
		Subrole:     cStringToGo(metadata.subrole),
		Title:       cStringToGo(metadata.title),
		Description: cStringToGo(metadata.description),
		Value:       cStringToGo(metadata.value),
		Enabled:     metadata.enabled != 0,
		Focused:     metadata.focused != 0,
		Frame:       common.Rect{X: int(metadata.frame.x), Y: int(metadata.frame.y), Width: int(metadata.frame.width), Height: int(metadata.frame.height)},
		FrameKnown:  metadata.has_frame != 0,
		Actions:     actions,
	}
}

func convertAXElementMatchList(matches C.gut_ax_element_match_list) []common.AXElementMatch {
	result := make([]common.AXElementMatch, 0, int(matches.length))
	if matches.length == 0 || matches.items == nil {
		return result
	}
	values := unsafe.Slice((*C.gut_ax_element_match)(unsafe.Pointer(matches.items)), int(matches.length))
	for _, match := range values {
		result = append(result, convertAXElementMatch(match))
	}
	return result
}

func convertAXElementMatch(match C.gut_ax_element_match) common.AXElementMatch {
	return common.AXElementMatch{
		Ref:              convertAXElementRef(match.ref),
		Metadata:         convertElementMetadata(match.metadata),
		Depth:            int(match.depth),
		ActionPoint:      common.Point{X: int(match.action_point.x), Y: int(match.action_point.y)},
		ActionPointKnown: match.has_action_point != 0,
	}
}

func convertAXElementRef(ref C.gut_ax_element_ref) common.AXElementRef {
	path := make([]int, 0, int(ref.path.length))
	if ref.path.length > 0 && ref.path.items != nil {
		values := unsafe.Slice((*C.int64_t)(unsafe.Pointer(ref.path.items)), int(ref.path.length))
		for _, value := range values {
			path = append(path, int(value))
		}
	}
	return common.AXElementRef{
		Scope:        common.AXSearchScope(cStringToGo(ref.scope)),
		OwnerPID:     int(ref.owner_pid),
		WindowHandle: common.WindowHandle(ref.window_handle),
		Path:         path,
	}
}

func makeAXElementSearchQuery(query common.AXElementSearchQuery) (C.gut_ax_element_search_query, func()) {
	cQuery := C.gut_ax_element_search_query{
		scope:                makeCStringOrNil(string(query.Scope)),
		window_handle:        C.int64_t(query.WindowHandle),
		role:                 makeCStringOrNil(query.Role),
		subrole:              makeCStringOrNil(query.Subrole),
		title_contains:       makeCStringOrNil(query.TitleContains),
		value_contains:       makeCStringOrNil(query.ValueContains),
		description_contains: makeCStringOrNil(query.DescriptionContains),
		action:               makeCStringOrNil(query.Action),
		enabled_state:        boolPointerToTristate(query.Enabled),
		focused_state:        boolPointerToTristate(query.Focused),
		limit:                C.int64_t(query.Limit),
		max_depth:            C.int64_t(query.MaxDepth),
	}
	return cQuery, func() {
		freeCString(cQuery.scope)
		freeCString(cQuery.role)
		freeCString(cQuery.subrole)
		freeCString(cQuery.title_contains)
		freeCString(cQuery.value_contains)
		freeCString(cQuery.description_contains)
		freeCString(cQuery.action)
	}
}

func makeAXElementRef(ref common.AXElementRef) (C.gut_ax_element_ref, func()) {
	cRef := C.gut_ax_element_ref{
		scope:         makeCStringOrNil(string(ref.Scope)),
		owner_pid:     C.int64_t(ref.OwnerPID),
		window_handle: C.int64_t(ref.WindowHandle),
	}
	if len(ref.Path) > 0 {
		items := C.malloc(C.size_t(len(ref.Path)) * C.size_t(unsafe.Sizeof(C.int64_t(0))))
		if items != nil {
			slice := unsafe.Slice((*C.int64_t)(items), len(ref.Path))
			for index, value := range ref.Path {
				slice[index] = C.int64_t(value)
			}
			cRef.path.items = (*C.int64_t)(items)
			cRef.path.length = C.int64_t(len(ref.Path))
		}
	}
	return cRef, func() {
		freeCString(cRef.scope)
		if cRef.path.items != nil {
			C.free(unsafe.Pointer(cRef.path.items))
		}
	}
}

func makeCStringOrNil(value string) *C.char {
	if value == "" {
		return nil
	}
	return C.CString(value)
}

func freeCString(value *C.char) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}

func boolPointerToTristate(value *bool) C.int {
	if value == nil {
		return -1
	}
	if *value {
		return 1
	}
	return 0
}

func validateAXElementSearchQuery(query common.AXElementSearchQuery) error {
	switch query.Scope {
	case common.AXSearchScopeFocusedWindow, common.AXSearchScopeFrontmostApplication, common.AXSearchScopeWindowHandle:
	default:
		return fmt.Errorf("%w: searchAXElements [%s]", common.ErrInvalidToken, common.CapabilityAXElementSearch)
	}
	if query.Scope == common.AXSearchScopeWindowHandle && query.WindowHandle == 0 {
		return fmt.Errorf("%w: searchAXElements [%s]", common.ErrInvalidToken, common.CapabilityAXElementSearch)
	}
	if query.Limit <= 0 {
		return fmt.Errorf("%w: searchAXElements [%s]", common.ErrInvalidToken, common.CapabilityAXElementSearch)
	}
	if query.MaxDepth < 0 {
		return fmt.Errorf("%w: searchAXElements [%s]", common.ErrInvalidToken, common.CapabilityAXElementSearch)
	}
	return nil
}

func validateAXElementRef(ref common.AXElementRef, operation string, capability common.Capability) error {
	switch ref.Scope {
	case common.AXSearchScopeFocusedWindow, common.AXSearchScopeFrontmostApplication, common.AXSearchScopeWindowHandle:
	default:
		return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
	}
	if ref.Scope == common.AXSearchScopeWindowHandle && ref.WindowHandle == 0 {
		return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
	}
	for _, index := range ref.Path {
		if index < 0 {
			return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
		}
	}
	return nil
}

func cStringToGo(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

func validateAXAction(action common.AXAction, operation string, capability common.Capability) error {
	if action == "" {
		return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
	}
	return nil
}

func bridgeAXInteractionError(operation string, capability common.Capability, status int, unavailableDetail string) error {
	switch status {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
	case 3:
		return common.UnsupportedOperation(operation, runtime.GOOS, capability, "operation is not supported by the linked native backend on this platform")
	case 4, 6:
		return common.CapabilityUnavailable(operation, runtime.GOOS, capability, unavailableDetail)
	case 5:
		return common.PermissionDeniedOperation(operation, runtime.GOOS, capability, "operation requires macOS Accessibility permission")
	default:
		return common.UnavailableOperation(operation, runtime.GOOS, capability, "linked native bridge call failed")
	}
}

func bridgeStatusError(operation string, capability common.Capability, status int) error {
	switch status {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("%w: %s [%s]", common.ErrInvalidToken, operation, capability)
	case 3:
		return common.UnsupportedOperation(operation, runtime.GOOS, capability, "operation is not supported by the linked native backend on this platform")
	case 4:
		return common.CapabilityUnavailable(operation, runtime.GOOS, capability, "operation is intentionally disabled by the linked native backend safety model")
	case 6:
		return common.CapabilityUnavailable(operation, runtime.GOOS, capability, "no AX element is available at the requested point")
	case 5:
		return common.PermissionDeniedOperation(operation, runtime.GOOS, capability, "operation requires macOS Accessibility permission")
	default:
		return common.UnavailableOperation(operation, runtime.GOOS, capability, "linked native bridge call failed")
	}
}
