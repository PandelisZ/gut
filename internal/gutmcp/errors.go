package gutmcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/PandelisZ/gut/imageproc"
	"github.com/PandelisZ/gut/native/common"
	"github.com/PandelisZ/gut/provider"
)

func actionableError(action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	var opErr *common.OperationError
	if errors.As(err, &opErr) {
		switch {
		case errors.Is(err, common.ErrPermissionDenied):
			return fmt.Errorf("%s failed: permission denied for capability %q. Grant the required desktop permission and retry: %w", action, opErr.Capability, err)
		case errors.Is(err, common.ErrCapabilityUnavailable):
			return fmt.Errorf("%s failed: capability %q is unavailable on this host: %w", action, opErr.Capability, err)
		case errors.Is(err, common.ErrUnsupportedPlatform):
			return fmt.Errorf("%s failed: this host does not support capability %q: %w", action, opErr.Capability, err)
		case errors.Is(err, common.ErrInvalidToken):
			return fmt.Errorf("%s failed: the provided selector or token is invalid: %w", action, err)
		}
	}

	var missing provider.MissingProviderError
	if errors.As(err, &missing) {
		return fmt.Errorf("%s failed: provider %q is not registered in this build", action, missing.Name)
	}

	switch {
	case errors.Is(err, imageproc.ErrWindowFinderUnavailable):
		return fmt.Errorf("%s failed: window matching is unavailable in the default gut registry", action)
	case errors.Is(err, imageproc.ErrTextFinderUnavailable):
		return fmt.Errorf("%s failed: text finding is unavailable in the default gut registry", action)
	case errors.Is(err, imageproc.ErrImageFinderUnavailable):
		return fmt.Errorf("%s failed: image matching is unavailable in the default gut registry", action)
	default:
		return fmt.Errorf("%s failed: %w", action, err)
	}
}
