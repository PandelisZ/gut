//go:build !darwin || !cgo

package libnutcore

func ShowAgentCursor(AgentCursorEvent) error {
	return nil
}

func HideAgentCursor() error {
	return nil
}
