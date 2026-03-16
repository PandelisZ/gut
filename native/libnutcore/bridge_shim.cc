#include "bridge_shim.h"
#include "bridge_shim_common.h"

#include <stdlib.h>

extern "C" {

int gut_drag_mouse(int64_t x, int64_t y, const char *button_token) {
	MMMouseButton button = LEFT_BUTTON;
	int status = gut_parse_mouse_button(button_token, &button);
	if (status != gut_status_ok) {
		return status;
	}
	dragMouse(MMPointMake(x, y), button);
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_move_mouse(int64_t x, int64_t y) {
	moveMouse(MMPointMake(x, y));
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_get_mouse_pos(gut_point *point) {
	if (point == NULL) {
		return gut_status_failed;
	}
	MMPoint native_point = getMousePos();
	point->x = native_point.x;
	point->y = native_point.y;
	return gut_status_ok;
}

int gut_mouse_click(const char *button_token, int double_click) {
	MMMouseButton button = LEFT_BUTTON;
	int status = gut_parse_mouse_button(button_token, &button);
	if (status != gut_status_ok) {
		return status;
	}
	if (double_click) {
		doubleClick(button);
	} else {
		clickMouse(button);
	}
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_mouse_toggle(const char *state_token, const char *button_token) {
	int down = 0;
	int status = gut_parse_button_state(state_token, &down);
	if (status != gut_status_ok) {
		return status;
	}
	MMMouseButton button = LEFT_BUTTON;
	status = gut_parse_mouse_button(button_token, &button);
	if (status != gut_status_ok) {
		return status;
	}
	toggleMouse(down != 0, button);
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_scroll_mouse(int horizontal, int vertical) {
	scrollMouse(horizontal, vertical);
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_set_mouse_delay(int64_t delay_ms) {
	gut_mouse_delay_ms = delay_ms;
	return gut_status_ok;
}

int gut_key_tap(const char *key_token, const char *const *modifier_tokens, size_t modifier_count) {
	MMKeyCode key;
	int status = gut_parse_key_code(key_token, &key);
	if (status != gut_status_ok) {
		return status;
	}
	MMKeyFlags flags = MOD_NONE;
	status = gut_collect_modifiers(modifier_tokens, modifier_count, &flags);
	if (status != gut_status_ok) {
		return status;
	}
	status = gut_validate_key_action(key, flags);
	if (status != gut_status_ok) {
		return status;
	}
	tapKeyCode(key, flags);
	gut_apply_keyboard_delay();
	return gut_status_ok;
}

int gut_key_toggle(const char *key_token, const char *state_token, const char *const *modifier_tokens, size_t modifier_count) {
	MMKeyCode key;
	int status = gut_parse_key_code(key_token, &key);
	if (status != gut_status_ok) {
		return status;
	}
	int down = 0;
	status = gut_parse_button_state(state_token, &down);
	if (status != gut_status_ok) {
		return status;
	}
	MMKeyFlags flags = MOD_NONE;
	status = gut_collect_modifiers(modifier_tokens, modifier_count, &flags);
	if (status != gut_status_ok) {
		return status;
	}
	status = gut_validate_key_action(key, flags);
	if (status != gut_status_ok) {
		return status;
	}
	toggleKeyCode(key, down != 0, flags);
	gut_apply_keyboard_delay();
	return gut_status_ok;
}

int gut_type_string(const char *text) {
	typeString(text);
	gut_apply_keyboard_delay();
	return gut_status_ok;
}

int gut_type_string_delayed(const char *text, int cpm) {
	typeStringDelayed(text, (unsigned int)cpm);
	return gut_status_ok;
}

int gut_set_keyboard_delay(int64_t delay_ms) {
	gut_keyboard_delay_ms = delay_ms;
	return gut_status_ok;
}

int gut_get_screen_size(gut_size *size) {
	if (size == NULL) {
		return gut_status_failed;
	}
	MMSize native_size = getMainDisplaySize();
	size->width = native_size.width;
	size->height = native_size.height;
	return gut_status_ok;
}

int gut_highlight(int64_t x, int64_t y, int64_t width, int64_t height, int64_t duration_ms, double opacity) {
	highlight((int32_t)x, (int32_t)y, (int32_t)width, (int32_t)height, (long)duration_ms, (float)opacity);
	return gut_status_ok;
}

int gut_capture_screen(const gut_rect *region, gut_bitmap **bitmap) {
	if (bitmap == NULL) {
		return gut_status_failed;
	}
#if defined(IS_MACOSX)
	(void)region;
	*bitmap = NULL;
	return gut_status_capability_unavailable;
#else
	MMRect native_region;
	if (region != NULL) {
		native_region = MMRectMake(region->x, region->y, region->width, region->height);
	} else {
		MMSize display_size = getMainDisplaySize();
		native_region = MMRectMake(0, 0, display_size.width, display_size.height);
	}
	MMBitmapRef native_bitmap = copyMMBitmapFromDisplayInRect(native_region);
	if (native_bitmap == NULL) {
		return gut_status_failed;
	}
	gut_bitmap *result = (gut_bitmap *)malloc(sizeof(gut_bitmap));
	if (result == NULL) {
		destroyMMBitmap(native_bitmap);
		return gut_status_failed;
	}
	size_t image_length = native_bitmap->bytewidth * native_bitmap->height;
	result->image = (unsigned char *)malloc(image_length);
	if (result->image == NULL) {
		free(result);
		destroyMMBitmap(native_bitmap);
		return gut_status_failed;
	}
	memcpy(result->image, native_bitmap->imageBuffer, image_length);
	result->width = native_bitmap->width;
	result->height = native_bitmap->height;
	result->byte_width = native_bitmap->bytewidth;
	result->bits_per_pixel = native_bitmap->bitsPerPixel;
	result->bytes_per_pixel = native_bitmap->bytesPerPixel;
	result->image_len = (int64_t)image_length;
	*bitmap = result;
	destroyMMBitmap(native_bitmap);
	return gut_status_ok;
#endif
}

int gut_get_windows(gut_window_list *windows) {
	if (windows == NULL) {
		return gut_status_failed;
	}
	std::vector<WindowHandle> native_windows = getWindows();
	windows->length = (int64_t)native_windows.size();
	windows->handles = NULL;
	if (native_windows.empty()) {
		return gut_status_ok;
	}
	windows->handles = (int64_t *)malloc(sizeof(int64_t) * native_windows.size());
	if (windows->handles == NULL) {
		windows->length = 0;
		return gut_status_failed;
	}
	for (size_t index = 0; index < native_windows.size(); index++) {
		windows->handles[index] = native_windows[index];
	}
	return gut_status_ok;
}

int gut_get_active_window(int64_t *window_handle) {
	if (window_handle == NULL) {
		return gut_status_failed;
	}
	*window_handle = getActiveWindow();
	return gut_status_ok;
}

int gut_get_window_rect(int64_t window_handle, gut_rect *rect) {
	if (rect == NULL) {
		return gut_status_failed;
	}
	MMRect native_rect = getWindowRect(window_handle);
	rect->x = native_rect.origin.x;
	rect->y = native_rect.origin.y;
	rect->width = native_rect.size.width;
	rect->height = native_rect.size.height;
	return gut_status_ok;
}

int gut_get_window_title(int64_t window_handle, char **title) {
	if (title == NULL) {
		return gut_status_failed;
	}
	std::string native_title = getWindowTitle(window_handle);
	*title = gut_copy_c_string(native_title.c_str());
	return *title == NULL ? gut_status_failed : gut_status_ok;
}

int gut_focus_window(int64_t window_handle, int *result) {
	if (result == NULL) {
		return gut_status_failed;
	}
	*result = focusWindow(window_handle) ? 1 : 0;
	return gut_status_ok;
}

int gut_move_window(int64_t window_handle, int64_t x, int64_t y, int *result) {
	if (result == NULL) {
		return gut_status_failed;
	}
	*result = moveWindow(window_handle, MMPointMake(x, y)) ? 1 : 0;
	return gut_status_ok;
}

int gut_resize_window(int64_t window_handle, int64_t width, int64_t height, int *result) {
	if (result == NULL) {
		return gut_status_failed;
	}
	*result = resizeWindow(window_handle, MMSizeMake(width, height)) ? 1 : 0;
	return gut_status_ok;
}

int gut_get_x_display_name(char **name) {
#if defined(USE_X11)
	if (name == NULL) {
		return gut_status_failed;
	}
	const char *current = getXDisplay();
	*name = gut_copy_c_string(current == NULL ? "" : current);
	return *name == NULL ? gut_status_failed : gut_status_ok;
#else
	(void)name;
	return gut_status_unsupported;
#endif
}

int gut_set_x_display_name(const char *name) {
#if defined(USE_X11)
	setXDisplay(name == NULL ? "" : name);
	return gut_status_ok;
#else
	(void)name;
	return gut_status_unsupported;
#endif
}

void gut_free_string(char *value) {
	free(value);
}

void gut_free_bitmap(gut_bitmap *bitmap) {
	if (bitmap == NULL) {
		return;
	}
	free(bitmap->image);
	free(bitmap);
}

void gut_free_window_list(gut_window_list *windows) {
	if (windows == NULL) {
		return;
	}
	free(windows->handles);
	windows->handles = NULL;
	windows->length = 0;
}

}
