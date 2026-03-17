//go:build darwin && cgo

package libnutcore

import "gut/native/common"

func newClient(options Options) Client {
	capabilities := linkedCapabilities()
	permissions, err := bridgeGetPermissionSnapshot()
	if err != nil {
		permissions = common.PermissionSnapshot{
			Accessibility: common.PermissionStatus{
				Supported: true,
				Reason:    "Accessibility permission state could not be determined from the native bridge",
			},
			ScreenRecording: common.PermissionStatus{
				Reason: "Screen Recording permission state could not be determined from the native bridge",
			},
		}
	}

	capabilities[common.CapabilityPermissionReadiness] = common.CapabilityStatus{
		Capability:   common.CapabilityPermissionReadiness,
		Availability: common.AvailabilityAvailable,
		Reason:       "macOS permission readiness is introspected with AXIsProcessTrustedWithOptions(prompt=false) and screen capture preflight when available",
	}

	if permissions.Accessibility.Granted {
		capabilities[common.CapabilityMouseMove] = common.CapabilityStatus{
			Capability:   common.CapabilityMouseMove,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native mouse bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityMouseDrag] = common.CapabilityStatus{
			Capability:   common.CapabilityMouseDrag,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native mouse bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityMouseClick] = common.CapabilityStatus{
			Capability:   common.CapabilityMouseClick,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native mouse bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityMouseToggle] = common.CapabilityStatus{
			Capability:   common.CapabilityMouseToggle,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native mouse bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityMouseScroll] = common.CapabilityStatus{
			Capability:   common.CapabilityMouseScroll,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native mouse bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityKeyboardTap] = common.CapabilityStatus{
			Capability:   common.CapabilityKeyboardTap,
			Availability: common.AvailabilityAvailable,
			Reason:       "macOS exposes primitive CGEvent-backed key taps only; modifier chords and special system-defined keys are rejected by the native safety model",
		}
		capabilities[common.CapabilityKeyboardToggle] = common.CapabilityStatus{
			Capability:   common.CapabilityKeyboardToggle,
			Availability: common.AvailabilityAvailable,
			Reason:       "macOS exposes primitive CGEvent-backed key toggles only; modifier chords and special system-defined keys are rejected by the native safety model",
		}
		capabilities[common.CapabilityKeyboardType] = common.CapabilityStatus{
			Capability:   common.CapabilityKeyboardType,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core native keyboard bridge is linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityWindowFocus] = common.CapabilityStatus{
			Capability:   common.CapabilityWindowFocus,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core window actions are linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityWindowMove] = common.CapabilityStatus{
			Capability:   common.CapabilityWindowMove,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core window actions are linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityWindowResize] = common.CapabilityStatus{
			Capability:   common.CapabilityWindowResize,
			Availability: common.AvailabilityAvailable,
			Reason:       "libnut-core window actions are linked via cgo and Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXFocusedWindowMetadata] = common.CapabilityStatus{
			Capability:   common.CapabilityAXFocusedWindowMetadata,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX focused-window metadata is available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXFocusedElementMetadata] = common.CapabilityStatus{
			Capability:   common.CapabilityAXFocusedElementMetadata,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX focused-element metadata is available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXElementAtPointMetadata] = common.CapabilityStatus{
			Capability:   common.CapabilityAXElementAtPointMetadata,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX element-at-point metadata is available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXFocusedWindowRaise] = common.CapabilityStatus{
			Capability:   common.CapabilityAXFocusedWindowRaise,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX raise on the focused window is available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXFocusedElementAction] = common.CapabilityStatus{
			Capability:   common.CapabilityAXFocusedElementAction,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX actions on the focused element are available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXElementActionAtPoint] = common.CapabilityStatus{
			Capability:   common.CapabilityAXElementActionAtPoint,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX actions on elements at screen points are available because Accessibility permission is granted",
		}
		capabilities[common.CapabilityAXElementFocusAtPoint] = common.CapabilityStatus{
			Capability:   common.CapabilityAXElementFocusAtPoint,
			Availability: common.AvailabilityAvailable,
			Reason:       "AX focus at screen points is available because Accessibility permission is granted",
		}
	} else {
		reason := permissions.Accessibility.Reason
		if reason == "" {
			reason = "Accessibility permission is not granted"
		}
		capabilities[common.CapabilityMouseMove] = common.CapabilityStatus{Capability: common.CapabilityMouseMove, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityMouseDrag] = common.CapabilityStatus{Capability: common.CapabilityMouseDrag, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityMouseClick] = common.CapabilityStatus{Capability: common.CapabilityMouseClick, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityMouseToggle] = common.CapabilityStatus{Capability: common.CapabilityMouseToggle, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityMouseScroll] = common.CapabilityStatus{Capability: common.CapabilityMouseScroll, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityKeyboardTap] = common.CapabilityStatus{Capability: common.CapabilityKeyboardTap, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityKeyboardToggle] = common.CapabilityStatus{Capability: common.CapabilityKeyboardToggle, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityKeyboardType] = common.CapabilityStatus{Capability: common.CapabilityKeyboardType, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityWindowFocus] = common.CapabilityStatus{Capability: common.CapabilityWindowFocus, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityWindowMove] = common.CapabilityStatus{Capability: common.CapabilityWindowMove, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityWindowResize] = common.CapabilityStatus{Capability: common.CapabilityWindowResize, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXFocusedWindowMetadata] = common.CapabilityStatus{Capability: common.CapabilityAXFocusedWindowMetadata, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXFocusedElementMetadata] = common.CapabilityStatus{Capability: common.CapabilityAXFocusedElementMetadata, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXElementAtPointMetadata] = common.CapabilityStatus{Capability: common.CapabilityAXElementAtPointMetadata, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXFocusedWindowRaise] = common.CapabilityStatus{Capability: common.CapabilityAXFocusedWindowRaise, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXFocusedElementAction] = common.CapabilityStatus{Capability: common.CapabilityAXFocusedElementAction, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXElementActionAtPoint] = common.CapabilityStatus{Capability: common.CapabilityAXElementActionAtPoint, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
		capabilities[common.CapabilityAXElementFocusAtPoint] = common.CapabilityStatus{Capability: common.CapabilityAXElementFocusAtPoint, Availability: common.AvailabilityPermissionBlocked, Reason: reason}
	}

	capabilities[common.CapabilityScreenHighlight] = common.CapabilityStatus{
		Capability:   common.CapabilityScreenHighlight,
		Availability: common.AvailabilityAvailable,
		Reason:       "macOS highlight overlays are dispatched onto the AppKit main thread by the local native shim",
	}
	capabilities[common.CapabilityScreenCapture] = common.CapabilityStatus{
		Capability:   common.CapabilityScreenCapture,
		Availability: common.AvailabilityUnsupported,
		Reason:       "macOS screen capture is intentionally disabled in the libnutcore safety model until a safe implementation is ready",
	}
	capabilities[common.CapabilityX11DisplayGet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplayGet,
		Availability: common.AvailabilityUnsupported,
		Reason:       "libnut-core X11 display access is only exposed on Linux",
	}
	capabilities[common.CapabilityX11DisplaySet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplaySet,
		Availability: common.AvailabilityUnsupported,
		Reason:       "libnut-core X11 display access is only exposed on Linux",
	}

	return newBridgeClient("darwin", []string{
		"libnut-core is linked directly through a local cgo shim on macOS.",
		"Quartz/CoreGraphics input primitives remain available, while AppKit overlay UI is dispatched onto the main thread inside libnutcore.",
		"macOS permission readiness is exposed without prompting so callers can distinguish unavailable capabilities from privacy-blocked ones.",
		"AX metadata and interaction primitives expose focused-window raise plus focused and point-targeted element actions without persistent AX handles.",
		"Screen capture is intentionally unavailable in the current macOS safety model and returns deterministic capability errors.",
	}, capabilities, options)
}
