//go:build linux && cgo

package libnutcore

import "gut/native/common"

func newClient(options Options) Client {
	capabilities := linkedCapabilities()
	capabilities[common.CapabilityX11DisplayGet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplayGet,
		Availability: common.AvailabilityAvailable,
		Reason:       "libnut-core native X11 display bridge is linked via cgo",
	}
	capabilities[common.CapabilityX11DisplaySet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplaySet,
		Availability: common.AvailabilityAvailable,
		Reason:       "libnut-core native X11 display bridge is linked via cgo",
	}

	return newBridgeClient("linux", []string{
		"libnut-core is linked directly through a local cgo shim on Linux.",
		"The linked backend uses the vendored X11/XTest implementation under libnut-core/src.",
	}, capabilities, options)
}
