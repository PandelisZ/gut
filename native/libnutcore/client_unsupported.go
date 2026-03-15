//go:build !windows && !linux && !darwin

package libnutcore

import "gut/native/common"

func newClient(options Options) Client {
	capabilities := unavailableCapabilities("libnut-core does not support this platform in the Go bridge")
	capabilities[common.CapabilityX11DisplayGet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplayGet,
		Availability: common.AvailabilityUnsupported,
		Reason:       "unsupported platform",
	}
	capabilities[common.CapabilityX11DisplaySet] = common.CapabilityStatus{
		Capability:   common.CapabilityX11DisplaySet,
		Availability: common.AvailabilityUnsupported,
		Reason:       "unsupported platform",
	}

	return newUnavailableClient("unsupported", []string{
		"No libnut-core backend exists for this platform in the Go bridge.",
	}, capabilities, options)
}
