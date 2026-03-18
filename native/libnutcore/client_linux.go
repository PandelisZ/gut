//go:build linux && !cgo

package libnutcore

import "github.com/PandelisZ/gut/native/common"

func newClient(options Options) Client {
	capabilities := unavailableCapabilities("cgo is disabled; rebuild with CGO_ENABLED=1 and a working libnut-core toolchain to enable native operations")
	capabilities[common.CapabilityX11DisplayGet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplayGet,
		Availability: common.AvailabilityAvailable,
		Reason:       "client-local X11 display configuration is available without native linkage",
	}
	capabilities[common.CapabilityX11DisplaySet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplaySet,
		Availability: common.AvailabilityAvailable,
		Reason:       "client-local X11 display configuration is available without native linkage",
	}

	return newUnavailableClient("linux", []string{
		"libnut-core native operations require cgo and Linux native development headers such as X11/XTest.",
		"This build keeps X11 display configuration local so higher layers can deterministically configure future native sessions.",
	}, capabilities, options)
}
