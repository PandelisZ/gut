//go:build windows && !cgo

package libnutcore

import "github.com/PandelisZ/gut/native/common"

func newClient(options Options) Client {
	capabilities := unavailableCapabilities("cgo is disabled; rebuild with CGO_ENABLED=1 and a Windows C/C++ toolchain to enable native operations")
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

	return newUnavailableClient("windows", []string{
		"libnut-core native operations require cgo plus a Windows C/C++ toolchain in this Go bridge.",
		"This default build reports deterministic unavailable capability reasons instead of silent stubs.",
	}, capabilities, options)
}
