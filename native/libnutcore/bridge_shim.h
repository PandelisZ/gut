#pragma once

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct gut_point {
	int64_t x;
	int64_t y;
} gut_point;

typedef struct gut_size {
	int64_t width;
	int64_t height;
} gut_size;

typedef struct gut_rect {
	int64_t x;
	int64_t y;
	int64_t width;
	int64_t height;
} gut_rect;

typedef struct gut_bitmap {
	int64_t width;
	int64_t height;
	int64_t byte_width;
	int64_t bits_per_pixel;
	int64_t bytes_per_pixel;
	int64_t image_len;
	unsigned char *image;
} gut_bitmap;

typedef struct gut_window_list {
	int64_t *handles;
	int64_t length;
} gut_window_list;

int gut_drag_mouse(int64_t x, int64_t y, const char *button_token);
int gut_move_mouse(int64_t x, int64_t y);
int gut_get_mouse_pos(gut_point *point);
int gut_mouse_click(const char *button_token, int double_click);
int gut_mouse_toggle(const char *state_token, const char *button_token);
int gut_scroll_mouse(int horizontal, int vertical);
int gut_set_mouse_delay(int64_t delay_ms);

int gut_key_tap(const char *key_token, const char *const *modifier_tokens, size_t modifier_count);
int gut_key_toggle(const char *key_token, const char *state_token, const char *const *modifier_tokens, size_t modifier_count);
int gut_type_string(const char *text);
int gut_type_string_delayed(const char *text, int cpm);
int gut_set_keyboard_delay(int64_t delay_ms);

int gut_get_screen_size(gut_size *size);
int gut_highlight(int64_t x, int64_t y, int64_t width, int64_t height, int64_t duration_ms, double opacity);
int gut_capture_screen(const gut_rect *region, gut_bitmap **bitmap);

int gut_get_windows(gut_window_list *windows);
int gut_get_active_window(int64_t *window_handle);
int gut_get_window_rect(int64_t window_handle, gut_rect *rect);
int gut_get_window_title(int64_t window_handle, char **title);
int gut_focus_window(int64_t window_handle, int *result);
int gut_move_window(int64_t window_handle, int64_t x, int64_t y, int *result);
int gut_resize_window(int64_t window_handle, int64_t width, int64_t height, int *result);

int gut_get_x_display_name(char **name);
int gut_set_x_display_name(const char *name);

void gut_free_string(char *value);
void gut_free_bitmap(gut_bitmap *bitmap);
void gut_free_window_list(gut_window_list *windows);

#ifdef __cplusplus
}
#endif
