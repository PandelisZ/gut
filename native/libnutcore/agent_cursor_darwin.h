#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

int gut_darwin_show_agent_cursor(
	const char *kind,
	int64_t position_x,
	int64_t position_y,
	int has_target,
	int64_t target_x,
	int64_t target_y,
	const char *button_token,
	const char *direction_token,
	int pressed,
	int64_t duration_ms
);

int gut_darwin_hide_agent_cursor(void);

#ifdef __cplusplus
}
#endif
