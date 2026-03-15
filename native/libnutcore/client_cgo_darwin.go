//go:build darwin && cgo

package libnutcore

import "gut/native/common"

func newClient(options Options) Client {
	capabilities := linkedCapabilities()
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
		"Native operations still depend on Accessibility and, for capture APIs, Screen Recording permission.",
	}, capabilities, options)
}
