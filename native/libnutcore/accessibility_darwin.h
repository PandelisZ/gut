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
int gut_darwin_search_ax_elements(const gut_ax_element_search_query *query, gut_ax_element_match_list *matches);
int gut_darwin_focus_ax_element(const gut_ax_element_ref *ref);
int gut_darwin_perform_ax_element_action(const gut_ax_element_ref *ref, const char *action_token);
void gut_darwin_free_window_metadata(gut_window_metadata *metadata);
void gut_darwin_free_element_metadata(gut_element_metadata *metadata);
void gut_darwin_free_ax_element_ref(gut_ax_element_ref *ref);
void gut_darwin_free_ax_element_match(gut_ax_element_match *match);
void gut_darwin_free_ax_element_match_list(gut_ax_element_match_list *matches);

#ifdef __cplusplus
}
#endif
