#pragma once

#include <stdlib.h>
#include <string.h>

#include "../../../libnut-core/src/types.h"
#include "../../../libnut-core/src/mouse.h"
#include "../../../libnut-core/src/keypress.h"
#include "../../../libnut-core/src/screen.h"
#include "../../../libnut-core/src/screengrab.h"
#include "../../../libnut-core/src/window_manager.h"
#include "../../../libnut-core/src/microsleep.h"
#if defined(USE_X11)
#include "../../../libnut-core/src/xdisplay.h"
#endif

enum {
	gut_status_ok = 0,
	gut_status_invalid_token = 1,
	gut_status_failed = 2,
	gut_status_unsupported = 3,
};

static int64_t gut_mouse_delay_ms = 10;
static int64_t gut_keyboard_delay_ms = 10;

typedef struct gut_key_name {
	const char *name;
	MMKeyCode key;
} gut_key_name;

static gut_key_name gut_key_names[] = {
	{"backspace", K_BACKSPACE},
	{"delete", K_DELETE},
	{"return", K_RETURN},
	{"tab", K_TAB},
	{"escape", K_ESCAPE},
	{"up", K_UP},
	{"down", K_DOWN},
	{"right", K_RIGHT},
	{"left", K_LEFT},
	{"home", K_HOME},
	{"end", K_END},
	{"pageup", K_PAGEUP},
	{"pagedown", K_PAGEDOWN},
	{"f1", K_F1},
	{"f2", K_F2},
	{"f3", K_F3},
	{"f4", K_F4},
	{"f5", K_F5},
	{"f6", K_F6},
	{"f7", K_F7},
	{"f8", K_F8},
	{"f9", K_F9},
	{"f10", K_F10},
	{"f11", K_F11},
	{"f12", K_F12},
	{"f13", K_F13},
	{"f14", K_F14},
	{"f15", K_F15},
	{"f16", K_F16},
	{"f17", K_F17},
	{"f18", K_F18},
	{"f19", K_F19},
	{"f20", K_F20},
	{"f21", K_F21},
	{"f22", K_F22},
	{"f23", K_F23},
	{"f24", K_F24},
	{"meta", K_META},
	{"right_meta", K_RIGHTMETA},
	{"cmd", K_CMD},
	{"right_cmd", K_RIGHTCMD},
	{"win", K_WIN},
	{"right_win", K_RIGHTWIN},
	{"alt", K_ALT},
	{"right_alt", K_RIGHTALT},
	{"control", K_CONTROL},
	{"right_control", K_RIGHTCONTROL},
	{"shift", K_SHIFT},
	{"right_shift", K_RIGHTSHIFT},
	{"space", K_SPACE},
	{"printscreen", K_PRINTSCREEN},
	{"insert", K_INSERT},
	{"menu", K_MENU},
	{"fn", K_FUNCTION},
	{"pause", K_PAUSE},
	{"caps_lock", K_CAPSLOCK},
	{"num_lock", K_NUMLOCK},
	{"scroll_lock", K_SCROLL_LOCK},
	{"audio_mute", K_AUDIO_VOLUME_MUTE},
	{"audio_vol_down", K_AUDIO_VOLUME_DOWN},
	{"audio_vol_up", K_AUDIO_VOLUME_UP},
	{"audio_play", K_AUDIO_PLAY},
	{"audio_stop", K_AUDIO_STOP},
	{"audio_pause", K_AUDIO_PAUSE},
	{"audio_prev", K_AUDIO_PREV},
	{"audio_next", K_AUDIO_NEXT},
	{"audio_rewind", K_AUDIO_REWIND},
	{"audio_forward", K_AUDIO_FORWARD},
	{"audio_repeat", K_AUDIO_REPEAT},
	{"audio_random", K_AUDIO_RANDOM},
	{"numpad_0", K_NUMPAD_0},
	{"numpad_1", K_NUMPAD_1},
	{"numpad_2", K_NUMPAD_2},
	{"numpad_3", K_NUMPAD_3},
	{"numpad_4", K_NUMPAD_4},
	{"numpad_5", K_NUMPAD_5},
	{"numpad_6", K_NUMPAD_6},
	{"numpad_7", K_NUMPAD_7},
	{"numpad_8", K_NUMPAD_8},
	{"numpad_9", K_NUMPAD_9},
	{"numpad_decimal", K_NUMPAD_DECIMAL},
	{"enter", K_ENTER},
	{"clear", K_CLEAR},
	{"add", K_ADD},
	{"subtract", K_SUBTRACT},
	{"multiply", K_MULTIPLY},
	{"divide", K_DIVIDE},
	{"lights_mon_up", K_LIGHTS_MON_UP},
	{"lights_mon_down", K_LIGHTS_MON_DOWN},
	{"lights_kbd_toggle", K_LIGHTS_KBD_TOGGLE},
	{"lights_kbd_up", K_LIGHTS_KBD_UP},
	{"lights_kbd_down", K_LIGHTS_KBD_DOWN},
	{NULL, K_NOT_A_KEY},
};

static int gut_parse_mouse_button(const char *button_token, MMMouseButton *button) {
	if (button_token == NULL || button == NULL) {
		return gut_status_invalid_token;
	}
	if (strcmp(button_token, "left") == 0) {
		*button = LEFT_BUTTON;
		return gut_status_ok;
	}
	if (strcmp(button_token, "right") == 0) {
		*button = RIGHT_BUTTON;
		return gut_status_ok;
	}
	if (strcmp(button_token, "middle") == 0) {
		*button = CENTER_BUTTON;
		return gut_status_ok;
	}
	return gut_status_invalid_token;
}

static int gut_parse_button_state(const char *state_token, int *down) {
	if (state_token == NULL || down == NULL) {
		return gut_status_invalid_token;
	}
	if (strcmp(state_token, "down") == 0) {
		*down = 1;
		return gut_status_ok;
	}
	if (strcmp(state_token, "up") == 0) {
		*down = 0;
		return gut_status_ok;
	}
	return gut_status_invalid_token;
}

static int gut_parse_key_code(const char *key_token, MMKeyCode *key) {
	if (key_token == NULL || key == NULL) {
		return gut_status_invalid_token;
	}
	if (strlen(key_token) == 1) {
		*key = keyCodeForChar(*key_token);
		return gut_status_ok;
	}
	for (gut_key_name *current = gut_key_names; current->name != NULL; current++) {
		if (strcmp(current->name, key_token) == 0) {
			*key = current->key;
			return gut_status_ok;
		}
	}
	return gut_status_invalid_token;
}

static int gut_parse_modifier_token(const char *modifier_token, MMKeyFlags *flags) {
	if (modifier_token == NULL || flags == NULL) {
		return gut_status_invalid_token;
	}
	if (strcmp(modifier_token, "alt") == 0 || strcmp(modifier_token, "right_alt") == 0) {
		*flags = MOD_ALT;
		return gut_status_ok;
	}
	if (strcmp(modifier_token, "control") == 0 || strcmp(modifier_token, "right_control") == 0) {
		*flags = MOD_CONTROL;
		return gut_status_ok;
	}
	if (strcmp(modifier_token, "shift") == 0 || strcmp(modifier_token, "right_shift") == 0) {
		*flags = MOD_SHIFT;
		return gut_status_ok;
	}
	if (strcmp(modifier_token, "fn") == 0) {
		*flags = MOD_FN;
		return gut_status_ok;
	}
	if (strcmp(modifier_token, "none") == 0) {
		*flags = MOD_NONE;
		return gut_status_ok;
	}
	if (
		strcmp(modifier_token, "meta") == 0 ||
		strcmp(modifier_token, "right_meta") == 0 ||
		strcmp(modifier_token, "command") == 0 ||
		strcmp(modifier_token, "cmd") == 0 ||
		strcmp(modifier_token, "right_cmd") == 0 ||
		strcmp(modifier_token, "win") == 0 ||
		strcmp(modifier_token, "right_win") == 0
	) {
		*flags = MOD_META;
		return gut_status_ok;
	}
	return gut_status_invalid_token;
}

static int gut_collect_modifiers(const char *const *modifier_tokens, size_t modifier_count, MMKeyFlags *flags) {
	if (flags == NULL) {
		return gut_status_invalid_token;
	}
	*flags = MOD_NONE;
	for (size_t index = 0; index < modifier_count; index++) {
		MMKeyFlags modifier = MOD_NONE;
		int status = gut_parse_modifier_token(modifier_tokens[index], &modifier);
		if (status != gut_status_ok) {
			return status;
		}
		*flags = (MMKeyFlags)(*flags | modifier);
	}
	return gut_status_ok;
}

static char *gut_copy_c_string(const char *value) {
	size_t length = value == NULL ? 0 : strlen(value);
	char *copy = (char *)malloc(length + 1);
	if (copy == NULL) {
		return NULL;
	}
	if (length > 0) {
		memcpy(copy, value, length);
	}
	copy[length] = '\0';
	return copy;
}

static void gut_apply_mouse_delay(void) {
	if (gut_mouse_delay_ms > 0) {
		microsleep(gut_mouse_delay_ms);
	}
}

static void gut_apply_keyboard_delay(void) {
	if (gut_keyboard_delay_ms > 0) {
		microsleep(gut_keyboard_delay_ms);
	}
}
