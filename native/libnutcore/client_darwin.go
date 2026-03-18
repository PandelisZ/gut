//go:build darwin && !cgo

package libnutcore

import "github.com/PandelisZ/gut/native/common"

func newClient(options Options) Client {
	capabilities := unavailableCapabilities("cgo is disabled; rebuild with CGO_ENABLED=1 and Apple system frameworks available to enable native operations")
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

	return newUnavailableClient("darwin", []string{
		"libnut-core native operations require cgo plus Objective-C compilation against ApplicationServices and Cocoa.",
		"Real mouse, keyboard, window, and screen capture operations also require Accessibility permission.",
		"Screen capture and some window metadata additionally require Screen Recording permission on recent macOS releases.",
	}, capabilities, options)
}
