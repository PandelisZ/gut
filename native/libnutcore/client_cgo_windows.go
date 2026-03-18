//go:build windows && cgo

package libnutcore

import "github.com/PandelisZ/gut/native/common"

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

	return newBridgeClient("windows", []string{
		"libnut-core is linked directly through a local cgo shim on Windows.",
		"The linked backend uses the vendored Win32 implementation under libnut-core/src.",
	}, capabilities, options)
}
