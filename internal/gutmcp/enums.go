package gutmcp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/shared"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

var keyNameToValue = buildKeyNameMap()

func buildKeyNameMap() map[string]shared.Key {
	result := make(map[string]shared.Key, int(shared.KeyAudioRandom)+1)
	for key := shared.Key(0); key <= shared.KeyAudioRandom; key++ {
		result[normalizeEnum(key.String())] = key
	}
	return result
}

func normalizeEnum(value string) string {
	return nonAlphaNum.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "")
}

func parseKey(name string) (shared.Key, error) {
	key, ok := keyNameToValue[normalizeEnum(name)]
	if !ok {
		return 0, fmt.Errorf("unknown key %q", name)
	}
	return key, nil
}

func canonicalKeyNames() []string {
	names := make([]string, 0, int(shared.KeyAudioRandom)+1)
	for key := shared.Key(0); key <= shared.KeyAudioRandom; key++ {
		names = append(names, key.String())
	}
	return names
}

func parseButton(name string) (shared.Button, error) {
	for button := shared.Button(0); button <= shared.ButtonRight; button++ {
		if normalizeEnum(button.String()) == normalizeEnum(name) {
			return button, nil
		}
	}
	return 0, fmt.Errorf("unknown mouse button %q", name)
}

func canonicalButtonNames() []string {
	names := make([]string, 0, int(shared.ButtonRight)+1)
	for button := shared.Button(0); button <= shared.ButtonRight; button++ {
		names = append(names, button.String())
	}
	return names
}

func parseAXAction(name string) (common.AXAction, error) {
	switch normalizeEnum(name) {
	case normalizeEnum(string(common.AXPress)):
		return common.AXPress, nil
	case normalizeEnum(string(common.AXRaise)):
		return common.AXRaise, nil
	case normalizeEnum(string(common.AXShowMenu)):
		return common.AXShowMenu, nil
	case normalizeEnum(string(common.AXConfirm)):
		return common.AXConfirm, nil
	case normalizeEnum(string(common.AXPick)):
		return common.AXPick, nil
	default:
		return "", fmt.Errorf("unknown accessibility action %q", name)
	}
}

func parseAXScope(name string) (common.AXSearchScope, error) {
	switch normalizeEnum(name) {
	case normalizeEnum(string(common.AXSearchScopeFocusedWindow)):
		return common.AXSearchScopeFocusedWindow, nil
	case normalizeEnum(string(common.AXSearchScopeFrontmostApplication)):
		return common.AXSearchScopeFrontmostApplication, nil
	case normalizeEnum(string(common.AXSearchScopeWindowHandle)):
		return common.AXSearchScopeWindowHandle, nil
	default:
		return "", fmt.Errorf("unknown accessibility search scope %q", name)
	}
}

func allCapabilities() []common.Capability {
	return []common.Capability{
		common.CapabilityMouseMove,
		common.CapabilityMouseDrag,
		common.CapabilityMousePosition,
		common.CapabilityMouseClick,
		common.CapabilityMouseToggle,
		common.CapabilityMouseScroll,
		common.CapabilityMouseDelay,
		common.CapabilityKeyboardTap,
		common.CapabilityKeyboardToggle,
		common.CapabilityKeyboardType,
		common.CapabilityKeyboardDelay,
		common.CapabilityScreenSize,
		common.CapabilityScreenHighlight,
		common.CapabilityScreenCapture,
		common.CapabilityWindowList,
		common.CapabilityWindowActive,
		common.CapabilityWindowRect,
		common.CapabilityWindowTitle,
		common.CapabilityWindowFocus,
		common.CapabilityWindowMove,
		common.CapabilityWindowResize,
		common.CapabilityWindowMinimize,
		common.CapabilityWindowRestore,
		common.CapabilityX11DisplayGet,
		common.CapabilityX11DisplaySet,
		common.CapabilityPermissionReadiness,
		common.CapabilityAXFocusedWindowMetadata,
		common.CapabilityAXFocusedElementMetadata,
		common.CapabilityAXElementAtPointMetadata,
		common.CapabilityAXFocusedWindowRaise,
		common.CapabilityAXFocusedElementAction,
		common.CapabilityAXElementActionAtPoint,
		common.CapabilityAXElementFocusAtPoint,
		common.CapabilityAXElementSearch,
		common.CapabilityAXElementFocusMatch,
		common.CapabilityAXElementActionMatch,
	}
}
