#include "bridge_shim.h"
#include "bridge_shim_common.h"

#if defined(IS_MACOSX)
#include "accessibility_darwin.h"
#include "agent_cursor_darwin.h"
#endif

#include <stdlib.h>

extern "C" {

#if defined(IS_MACOSX)
static int gut_require_accessibility_permission(void) {
	return gut_darwin_accessibility_permission_granted() ? gut_status_ok : gut_status_permission_denied;
}
#else
static int gut_require_accessibility_permission(void) {
	return gut_status_ok;
}
#endif

int gut_get_permission_snapshot(gut_permission_snapshot *snapshot) {
#if defined(IS_MACOSX)
	return gut_darwin_get_permission_snapshot(snapshot);
#else
	(void)snapshot;
	return gut_status_unsupported;
#endif
}

int gut_get_focused_window_metadata(gut_window_metadata *metadata) {
#if defined(IS_MACOSX)
	return gut_darwin_get_focused_window_metadata(metadata);
#else
	(void)metadata;
	return gut_status_unsupported;
#endif
}

int gut_raise_focused_window(void) {
#if defined(IS_MACOSX)
	return gut_darwin_raise_focused_window();
#else
	return gut_status_unsupported;
#endif
}

int gut_get_focused_element_metadata(gut_element_metadata *metadata) {
#if defined(IS_MACOSX)
	return gut_darwin_get_focused_element_metadata(metadata);
#else
	(void)metadata;
	return gut_status_unsupported;
#endif
}

int gut_perform_focused_element_action(const char *action_token) {
#if defined(IS_MACOSX)
	return gut_darwin_perform_focused_element_action(action_token);
#else
	(void)action_token;
	return gut_status_unsupported;
#endif
}

int gut_get_element_metadata_at_point(int64_t x, int64_t y, gut_element_metadata *metadata) {
#if defined(IS_MACOSX)
	return gut_darwin_get_element_metadata_at_point(x, y, metadata);
#else
	(void)x;
	(void)y;
	(void)metadata;
	return gut_status_unsupported;
#endif
}

int gut_perform_element_action_at_point(int64_t x, int64_t y, const char *action_token) {
#if defined(IS_MACOSX)
	return gut_darwin_perform_element_action_at_point(x, y, action_token);
#else
	(void)x;
	(void)y;
	(void)action_token;
	return gut_status_unsupported;
#endif
}

int gut_focus_element_at_point(int64_t x, int64_t y) {
#if defined(IS_MACOSX)
	return gut_darwin_focus_element_at_point(x, y);
#else
	(void)x;
	(void)y;
	return gut_status_unsupported;
#endif
}

int gut_search_ax_elements(const gut_ax_element_search_query *query, gut_ax_element_match_list *matches) {
#if defined(IS_MACOSX)
	return gut_darwin_search_ax_elements(query, matches);
#else
	(void)query;
	(void)matches;
	return gut_status_unsupported;
#endif
}

int gut_focus_ax_element(const gut_ax_element_ref *ref) {
#if defined(IS_MACOSX)
	return gut_darwin_focus_ax_element(ref);
#else
	(void)ref;
	return gut_status_unsupported;
#endif
}

int gut_perform_ax_element_action(const gut_ax_element_ref *ref, const char *action_token) {
#if defined(IS_MACOSX)
	return gut_darwin_perform_ax_element_action(ref, action_token);
#else
	(void)ref;
	(void)action_token;
	return gut_status_unsupported;
#endif
}

int gut_drag_mouse(int64_t x, int64_t y, const char *button_token) {
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
	scrollMouse(horizontal, vertical);
	gut_apply_mouse_delay();
	return gut_status_ok;
}

int gut_set_mouse_delay(int64_t delay_ms) {
	gut_mouse_delay_ms = delay_ms;
	return gut_status_ok;
}

int gut_key_tap(const char *key_token, const char *const *modifier_tokens, size_t modifier_count) {
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
	typeString(text);
	gut_apply_keyboard_delay();
	return gut_status_ok;
}

int gut_type_string_delayed(const char *text, int cpm) {
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
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

int gut_show_agent_cursor(const char *kind, int64_t position_x, int64_t position_y, int has_target, int64_t target_x, int64_t target_y, const char *button_token, const char *direction_token, int pressed, int64_t duration_ms) {
#if defined(IS_MACOSX)
	return gut_darwin_show_agent_cursor(kind, position_x, position_y, has_target, target_x, target_y, button_token, direction_token, pressed, duration_ms);
#else
	(void)kind;
	(void)position_x;
	(void)position_y;
	(void)has_target;
	(void)target_x;
	(void)target_y;
	(void)button_token;
	(void)direction_token;
	(void)pressed;
	(void)duration_ms;
	return gut_status_unsupported;
#endif
}

int gut_hide_agent_cursor(void) {
#if defined(IS_MACOSX)
	return gut_darwin_hide_agent_cursor();
#else
	return gut_status_unsupported;
#endif
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
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
	*result = focusWindow(window_handle) ? 1 : 0;
	return gut_status_ok;
}

int gut_move_window(int64_t window_handle, int64_t x, int64_t y, int *result) {
	if (result == NULL) {
		return gut_status_failed;
	}
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
	}
	*result = moveWindow(window_handle, MMPointMake(x, y)) ? 1 : 0;
	return gut_status_ok;
}

int gut_resize_window(int64_t window_handle, int64_t width, int64_t height, int *result) {
	if (result == NULL) {
		return gut_status_failed;
	}
	int permission_status = gut_require_accessibility_permission();
	if (permission_status != gut_status_ok) {
		return permission_status;
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

void gut_free_window_metadata(gut_window_metadata *metadata) {
#if defined(IS_MACOSX)
	gut_darwin_free_window_metadata(metadata);
#else
	(void)metadata;
#endif
}

void gut_free_element_metadata(gut_element_metadata *metadata) {
#if defined(IS_MACOSX)
	gut_darwin_free_element_metadata(metadata);
#else
	(void)metadata;
#endif
}

void gut_free_ax_element_ref(gut_ax_element_ref *ref) {
#if defined(IS_MACOSX)
	gut_darwin_free_ax_element_ref(ref);
#else
	(void)ref;
#endif
}

void gut_free_ax_element_match(gut_ax_element_match *match) {
#if defined(IS_MACOSX)
	gut_darwin_free_ax_element_match(match);
#else
	(void)match;
#endif
}

void gut_free_ax_element_match_list(gut_ax_element_match_list *matches) {
#if defined(IS_MACOSX)
	gut_darwin_free_ax_element_match_list(matches);
#else
	(void)matches;
#endif
}

}
