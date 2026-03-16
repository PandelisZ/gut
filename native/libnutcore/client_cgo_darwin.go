//go:build darwin && cgo

package libnutcore

import "gut/native/common"

func newClient(options Options) Client {
	capabilities := linkedCapabilities()
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
		"Screen capture is intentionally unavailable in the current macOS safety model and returns deterministic capability errors.",
		"Native operations still depend on Accessibility and, for some window metadata, Screen Recording permission.",
	}, capabilities, options)
}
