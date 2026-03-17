package provider

import (
	"context"

	"gut/native/common"
	"gut/shared"
)

type libnutcoreAccessibilityProvider struct {
	client libnutcoreClient
}

func NewLibnutcoreAccessibilityProvider(client libnutcoreClient) AccessibilityProvider {
	return &libnutcoreAccessibilityProvider{client: client}
}

func (p *libnutcoreAccessibilityProvider) GetPermissionSnapshot(ctx context.Context) (common.PermissionSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return common.PermissionSnapshot{}, err
	}
	return p.client.GetPermissionSnapshot()
}

func (p *libnutcoreAccessibilityProvider) GetFocusedWindow(ctx context.Context) (common.FocusedWindowMetadata, error) {
	if err := ctx.Err(); err != nil {
		return common.FocusedWindowMetadata{}, err
	}
	return p.client.GetFocusedWindow()
}

func (p *libnutcoreAccessibilityProvider) GetFocusedElement(ctx context.Context) (common.UIElementMetadata, error) {
	if err := ctx.Err(); err != nil {
		return common.UIElementMetadata{}, err
	}
	return p.client.GetFocusedElement()
}

func (p *libnutcoreAccessibilityProvider) GetElementAtPoint(ctx context.Context, point shared.Point) (common.UIElementMetadata, error) {
	if err := ctx.Err(); err != nil {
		return common.UIElementMetadata{}, err
	}
	return p.client.GetElementAtPoint(pointToNative(point))
}

func (p *libnutcoreAccessibilityProvider) RaiseFocusedWindow(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.RaiseFocusedWindow()
}

func (p *libnutcoreAccessibilityProvider) PerformFocusedElementAction(ctx context.Context, action common.AXAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.PerformFocusedElementAction(action)
}

func (p *libnutcoreAccessibilityProvider) PerformElementActionAtPoint(ctx context.Context, point shared.Point, action common.AXAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.PerformElementActionAtPoint(pointToNative(point), action)
}

func (p *libnutcoreAccessibilityProvider) FocusElementAtPoint(ctx context.Context, point shared.Point) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.FocusElementAtPoint(pointToNative(point))
}

func (p *libnutcoreAccessibilityProvider) Capabilities() common.CapabilitySet {
	return p.client.Capabilities()
}
