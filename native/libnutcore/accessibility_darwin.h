#pragma once

#include "bridge_shim.h"

#ifdef __cplusplus
extern "C" {
#endif

int gut_darwin_accessibility_permission_granted(void);
int gut_darwin_get_permission_snapshot(gut_permission_snapshot *snapshot);
int gut_darwin_get_focused_window_metadata(gut_window_metadata *metadata);
int gut_darwin_raise_focused_window(void);
int gut_darwin_get_focused_element_metadata(gut_element_metadata *metadata);
int gut_darwin_perform_focused_element_action(const char *action_token);
int gut_darwin_get_element_metadata_at_point(int64_t x, int64_t y, gut_element_metadata *metadata);
int gut_darwin_perform_element_action_at_point(int64_t x, int64_t y, const char *action_token);
int gut_darwin_focus_element_at_point(int64_t x, int64_t y);
void gut_darwin_free_window_metadata(gut_window_metadata *metadata);
void gut_darwin_free_element_metadata(gut_element_metadata *metadata);

#ifdef __cplusplus
}
#endif
