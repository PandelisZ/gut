package backgroundmouse

import "errors"

var (
	ErrUnsupportedPlatform = errors.New("background mouse is only supported on darwin")
	ErrUnresolved          = errors.New("background mouse target could not be resolved safely")
	ErrActionUnsupported   = errors.New("background mouse action is not supported for the resolved target")
)
