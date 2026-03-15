package common

import "fmt"

func ParseMouseButton(value string) (MouseButton, error) {
	switch MouseButton(value) {
	case MouseButtonLeft, MouseButtonRight, MouseButtonMiddle:
		return MouseButton(value), nil
	default:
		return "", fmt.Errorf("%w: mouse button %q", ErrInvalidToken, value)
	}
}

func ParseButtonState(value string) (ButtonState, error) {
	switch ButtonState(value) {
	case ButtonStateUp, ButtonStateDown:
		return ButtonState(value), nil
	default:
		return "", fmt.Errorf("%w: button state %q", ErrInvalidToken, value)
	}
}

func ParseKeyState(value string) (KeyState, error) {
	switch KeyState(value) {
	case KeyStateUp, KeyStateDown:
		return KeyState(value), nil
	default:
		return "", fmt.Errorf("%w: key state %q", ErrInvalidToken, value)
	}
}

func NormalizeModifierToken(platform, value string) (KeyModifier, error) {
	switch value {
	case "alt", "right_alt":
		return KeyModifierAlt, nil
	case "control", "right_control":
		return KeyModifierControl, nil
	case "shift", "right_shift":
		return KeyModifierShift, nil
	case "fn":
		return KeyModifierFn, nil
	case "none":
		return KeyModifierNone, nil
	case "command", "cmd", "right_cmd":
		if platform == "darwin" {
			return KeyModifierMeta, nil
		}
	case "win", "right_win":
		if platform != "darwin" {
			return KeyModifierMeta, nil
		}
	case "meta", "right_meta":
		return KeyModifierMeta, nil
	}
	return "", fmt.Errorf("%w: modifier %q on %s", ErrInvalidToken, value, platform)
}
