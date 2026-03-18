package provider

import (
	"fmt"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
)

func keyToLibnutToken(key shared.Key) (string, error) {
	switch key {
	case shared.KeyEscape:
		return "escape", nil
	case shared.KeyF1:
		return "f1", nil
	case shared.KeyF2:
		return "f2", nil
	case shared.KeyF3:
		return "f3", nil
	case shared.KeyF4:
		return "f4", nil
	case shared.KeyF5:
		return "f5", nil
	case shared.KeyF6:
		return "f6", nil
	case shared.KeyF7:
		return "f7", nil
	case shared.KeyF8:
		return "f8", nil
	case shared.KeyF9:
		return "f9", nil
	case shared.KeyF10:
		return "f10", nil
	case shared.KeyF11:
		return "f11", nil
	case shared.KeyF12:
		return "f12", nil
	case shared.KeyF13:
		return "f13", nil
	case shared.KeyF14:
		return "f14", nil
	case shared.KeyF15:
		return "f15", nil
	case shared.KeyF16:
		return "f16", nil
	case shared.KeyF17:
		return "f17", nil
	case shared.KeyF18:
		return "f18", nil
	case shared.KeyF19:
		return "f19", nil
	case shared.KeyF20:
		return "f20", nil
	case shared.KeyF21:
		return "f21", nil
	case shared.KeyF22:
		return "f22", nil
	case shared.KeyF23:
		return "f23", nil
	case shared.KeyF24:
		return "f24", nil
	case shared.KeyPrint:
		return "printscreen", nil
	case shared.KeyScrollLock:
		return "scroll_lock", nil
	case shared.KeyPause:
		return "pause", nil
	case shared.KeyGrave:
		return "`", nil
	case shared.KeyNum1:
		return "1", nil
	case shared.KeyNum2:
		return "2", nil
	case shared.KeyNum3:
		return "3", nil
	case shared.KeyNum4:
		return "4", nil
	case shared.KeyNum5:
		return "5", nil
	case shared.KeyNum6:
		return "6", nil
	case shared.KeyNum7:
		return "7", nil
	case shared.KeyNum8:
		return "8", nil
	case shared.KeyNum9:
		return "9", nil
	case shared.KeyNum0:
		return "0", nil
	case shared.KeyMinus:
		return "-", nil
	case shared.KeyEqual:
		return "=", nil
	case shared.KeyBackspace:
		return "backspace", nil
	case shared.KeyInsert:
		return "insert", nil
	case shared.KeyHome:
		return "home", nil
	case shared.KeyPageUp:
		return "pageup", nil
	case shared.KeyNumLock:
		return "num_lock", nil
	case shared.KeyDivide:
		return "divide", nil
	case shared.KeyMultiply:
		return "multiply", nil
	case shared.KeySubtract:
		return "subtract", nil
	case shared.KeyTab:
		return "tab", nil
	case shared.KeyQ:
		return "q", nil
	case shared.KeyW:
		return "w", nil
	case shared.KeyE:
		return "e", nil
	case shared.KeyR:
		return "r", nil
	case shared.KeyT:
		return "t", nil
	case shared.KeyY:
		return "y", nil
	case shared.KeyU:
		return "u", nil
	case shared.KeyI:
		return "i", nil
	case shared.KeyO:
		return "o", nil
	case shared.KeyP:
		return "p", nil
	case shared.KeyLeftBracket:
		return "[", nil
	case shared.KeyRightBracket:
		return "]", nil
	case shared.KeyBackslash:
		return "\\", nil
	case shared.KeyDelete:
		return "delete", nil
	case shared.KeyEnd:
		return "end", nil
	case shared.KeyPageDown:
		return "pagedown", nil
	case shared.KeyNumPad7:
		return "numpad_7", nil
	case shared.KeyNumPad8:
		return "numpad_8", nil
	case shared.KeyNumPad9:
		return "numpad_9", nil
	case shared.KeyAdd:
		return "add", nil
	case shared.KeyCapsLock:
		return "caps_lock", nil
	case shared.KeyA:
		return "a", nil
	case shared.KeyS:
		return "s", nil
	case shared.KeyD:
		return "d", nil
	case shared.KeyF:
		return "f", nil
	case shared.KeyG:
		return "g", nil
	case shared.KeyH:
		return "h", nil
	case shared.KeyJ:
		return "j", nil
	case shared.KeyK:
		return "k", nil
	case shared.KeyL:
		return "l", nil
	case shared.KeySemicolon:
		return ";", nil
	case shared.KeyQuote:
		return "'", nil
	case shared.KeyReturn:
		return "return", nil
	case shared.KeyNumPad4:
		return "numpad_4", nil
	case shared.KeyNumPad5:
		return "numpad_5", nil
	case shared.KeyNumPad6:
		return "numpad_6", nil
	case shared.KeyLeftShift:
		return "shift", nil
	case shared.KeyZ:
		return "z", nil
	case shared.KeyX:
		return "x", nil
	case shared.KeyC:
		return "c", nil
	case shared.KeyV:
		return "v", nil
	case shared.KeyB:
		return "b", nil
	case shared.KeyN:
		return "n", nil
	case shared.KeyM:
		return "m", nil
	case shared.KeyComma:
		return ",", nil
	case shared.KeyPeriod:
		return ".", nil
	case shared.KeySlash:
		return "/", nil
	case shared.KeyRightShift:
		return "right_shift", nil
	case shared.KeyUp:
		return "up", nil
	case shared.KeyNumPad1:
		return "numpad_1", nil
	case shared.KeyNumPad2:
		return "numpad_2", nil
	case shared.KeyNumPad3:
		return "numpad_3", nil
	case shared.KeyEnter:
		return "enter", nil
	case shared.KeyLeftControl:
		return "control", nil
	case shared.KeyLeftSuper:
		return "meta", nil
	case shared.KeyLeftWin:
		return "win", nil
	case shared.KeyLeftCmd:
		return "cmd", nil
	case shared.KeyLeftAlt:
		return "alt", nil
	case shared.KeySpace:
		return "space", nil
	case shared.KeyRightAlt:
		return "right_alt", nil
	case shared.KeyRightSuper:
		return "right_meta", nil
	case shared.KeyRightWin:
		return "right_win", nil
	case shared.KeyRightCmd:
		return "right_cmd", nil
	case shared.KeyMenu:
		return "menu", nil
	case shared.KeyRightControl:
		return "right_control", nil
	case shared.KeyFn:
		return "fn", nil
	case shared.KeyLeft:
		return "left", nil
	case shared.KeyDown:
		return "down", nil
	case shared.KeyRight:
		return "right", nil
	case shared.KeyNumPad0:
		return "numpad_0", nil
	case shared.KeyDecimal:
		return "numpad_decimal", nil
	case shared.KeyClear:
		return "clear", nil
	case shared.KeyAudioMute:
		return "audio_mute", nil
	case shared.KeyAudioVolDown:
		return "audio_vol_down", nil
	case shared.KeyAudioVolUp:
		return "audio_vol_up", nil
	case shared.KeyAudioPlay:
		return "audio_play", nil
	case shared.KeyAudioStop:
		return "audio_stop", nil
	case shared.KeyAudioPause:
		return "audio_pause", nil
	case shared.KeyAudioPrev:
		return "audio_prev", nil
	case shared.KeyAudioNext:
		return "audio_next", nil
	case shared.KeyAudioRewind:
		return "audio_rewind", nil
	case shared.KeyAudioForward:
		return "audio_forward", nil
	case shared.KeyAudioRepeat:
		return "audio_repeat", nil
	case shared.KeyAudioRandom:
		return "audio_random", nil
	default:
		return "", fmt.Errorf("%w: shared key %s", common.ErrInvalidToken, key)
	}
}

func modifierToLibnutToken(key shared.Key) (string, bool) {
	switch key {
	case shared.KeyLeftShift:
		return "shift", true
	case shared.KeyRightShift:
		return "right_shift", true
	case shared.KeyLeftControl:
		return "control", true
	case shared.KeyRightControl:
		return "right_control", true
	case shared.KeyLeftAlt:
		return "alt", true
	case shared.KeyRightAlt:
		return "right_alt", true
	case shared.KeyLeftSuper:
		return "meta", true
	case shared.KeyRightSuper:
		return "right_meta", true
	case shared.KeyLeftWin:
		return "win", true
	case shared.KeyRightWin:
		return "right_win", true
	case shared.KeyLeftCmd:
		return "cmd", true
	case shared.KeyRightCmd:
		return "right_cmd", true
	case shared.KeyFn:
		return "fn", true
	default:
		return "", false
	}
}

func buttonToNative(button shared.Button) (common.MouseButton, error) {
	switch button {
	case shared.ButtonLeft:
		return common.MouseButtonLeft, nil
	case shared.ButtonMiddle:
		return common.MouseButtonMiddle, nil
	case shared.ButtonRight:
		return common.MouseButtonRight, nil
	default:
		return "", fmt.Errorf("%w: shared button %s", common.ErrInvalidToken, button)
	}
}

func pointToNative(point shared.Point) common.Point {
	return common.Point{X: point.X, Y: point.Y}
}

func sizeToNative(size shared.Size) common.Size {
	return common.Size{Width: size.Width, Height: size.Height}
}

func regionToNative(region shared.Region) common.Rect {
	return common.Rect{X: region.Left, Y: region.Top, Width: region.Width, Height: region.Height}
}

func pointFromNative(point common.Point) shared.Point {
	return shared.Point{X: point.X, Y: point.Y}
}

func regionFromNative(rect common.Rect) shared.Region {
	return shared.Region{Left: rect.X, Top: rect.Y, Width: rect.Width, Height: rect.Height}
}

func windowHandleToNative(handle shared.WindowHandle) common.WindowHandle {
	return common.WindowHandle(handle)
}

func windowHandleFromNative(handle common.WindowHandle) shared.WindowHandle {
	return shared.WindowHandle(handle)
}
