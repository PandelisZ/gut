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

	"gut/native/common"
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
	default:
		return common.UnavailableOperation(operation, runtime.GOOS, capability, "linked native bridge call failed")
	}
}
