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

typedef struct gut_permission_snapshot {
	int accessibility_granted;
	int accessibility_supported;
	int screen_recording_granted;
	int screen_recording_supported;
} gut_permission_snapshot;

typedef struct gut_string_list {
	char **items;
	int64_t length;
} gut_string_list;

typedef struct gut_window_metadata {
	int64_t handle;
	char *title;
	char *role;
	char *subrole;
	gut_rect rect;
	int has_rect;
	int focused;
	int main;
	int minimized;
	int64_t owner_pid;
	char *owner_name;
	char *bundle_id;
} gut_window_metadata;

typedef struct gut_element_metadata {
	char *role;
	char *subrole;
	char *title;
	char *description;
	char *value;
	int enabled;
	int focused;
	gut_rect frame;
	int has_frame;
	gut_string_list actions;
} gut_element_metadata;

typedef struct gut_int64_list {
	int64_t *items;
	int64_t length;
} gut_int64_list;

typedef struct gut_ax_element_search_query {
	char *scope;
	int64_t window_handle;
	char *role;
	char *subrole;
	char *title_contains;
	char *value_contains;
	char *description_contains;
	char *action;
	int enabled_state;
	int focused_state;
	int64_t limit;
	int64_t max_depth;
} gut_ax_element_search_query;

typedef struct gut_ax_element_ref {
	char *scope;
	int64_t owner_pid;
	int64_t window_handle;
	gut_int64_list path;
} gut_ax_element_ref;

typedef struct gut_ax_element_match {
	gut_ax_element_ref ref;
	gut_element_metadata metadata;
	int64_t depth;
	gut_point action_point;
	int has_action_point;
} gut_ax_element_match;

typedef struct gut_ax_element_match_list {
	gut_ax_element_match *items;
	int64_t length;
} gut_ax_element_match_list;

typedef struct gut_window_list {
	int64_t *handles;
	int64_t length;
} gut_window_list;

int gut_get_permission_snapshot(gut_permission_snapshot *snapshot);
int gut_get_focused_window_metadata(gut_window_metadata *metadata);
int gut_raise_focused_window(void);
int gut_get_focused_element_metadata(gut_element_metadata *metadata);
int gut_perform_focused_element_action(const char *action_token);
int gut_get_element_metadata_at_point(int64_t x, int64_t y, gut_element_metadata *metadata);
int gut_perform_element_action_at_point(int64_t x, int64_t y, const char *action_token);
int gut_focus_element_at_point(int64_t x, int64_t y);
int gut_search_ax_elements(const gut_ax_element_search_query *query, gut_ax_element_match_list *matches);
int gut_focus_ax_element(const gut_ax_element_ref *ref);
int gut_perform_ax_element_action(const gut_ax_element_ref *ref, const char *action_token);
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
int gut_show_agent_cursor(const char *kind, int64_t position_x, int64_t position_y, int has_target, int64_t target_x, int64_t target_y, const char *button_token, const char *direction_token, int pressed, int64_t duration_ms);
int gut_hide_agent_cursor(void);
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
void gut_free_window_metadata(gut_window_metadata *metadata);
void gut_free_element_metadata(gut_element_metadata *metadata);
void gut_free_ax_element_ref(gut_ax_element_ref *ref);
void gut_free_ax_element_match(gut_ax_element_match *match);
void gut_free_ax_element_match_list(gut_ax_element_match_list *matches);

#ifdef __cplusplus
}
#endif
