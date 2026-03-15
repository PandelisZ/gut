package provider

import (
	"context"
	"time"

	"gut/native/common"
	"gut/shared"
)

type libnutcoreMouseProvider struct {
	client libnutcoreClient
}

func NewLibnutcoreMouseProvider(client libnutcoreClient) MouseProvider {
	return &libnutcoreMouseProvider{client: client}
}

func (p *libnutcoreMouseProvider) SetMouseDelay(delay time.Duration) {
	_ = p.client.SetMouseDelay(delay)
}

func (p *libnutcoreMouseProvider) SetMousePosition(ctx context.Context, point shared.Point) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.MoveMouse(pointToNative(point))
}

func (p *libnutcoreMouseProvider) CurrentMousePosition(ctx context.Context) (shared.Point, error) {
	if err := ctx.Err(); err != nil {
		return shared.Point{}, err
	}
	point, err := p.client.GetMousePosition()
	if err != nil {
		return shared.Point{}, err
	}
	return pointFromNative(point), nil
}

func (p *libnutcoreMouseProvider) Click(ctx context.Context, button shared.Button) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	nativeButton, err := buttonToNative(button)
	if err != nil {
		return err
	}
	return p.client.MouseClick(nativeButton, false)
}

func (p *libnutcoreMouseProvider) DoubleClick(ctx context.Context, button shared.Button) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	nativeButton, err := buttonToNative(button)
	if err != nil {
		return err
	}
	return p.client.MouseClick(nativeButton, true)
}

func (p *libnutcoreMouseProvider) ScrollUp(ctx context.Context, amount int) error {
	return p.scroll(ctx, 0, amount)
}

func (p *libnutcoreMouseProvider) ScrollDown(ctx context.Context, amount int) error {
	return p.scroll(ctx, 0, -amount)
}

func (p *libnutcoreMouseProvider) ScrollLeft(ctx context.Context, amount int) error {
	return p.scroll(ctx, -amount, 0)
}

func (p *libnutcoreMouseProvider) ScrollRight(ctx context.Context, amount int) error {
	return p.scroll(ctx, amount, 0)
}

func (p *libnutcoreMouseProvider) PressButton(ctx context.Context, button shared.Button) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	nativeButton, err := buttonToNative(button)
	if err != nil {
		return err
	}
	return p.client.MouseToggle(common.ButtonStateDown, nativeButton)
}

func (p *libnutcoreMouseProvider) ReleaseButton(ctx context.Context, button shared.Button) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	nativeButton, err := buttonToNative(button)
	if err != nil {
		return err
	}
	return p.client.MouseToggle(common.ButtonStateUp, nativeButton)
}

func (p *libnutcoreMouseProvider) scroll(ctx context.Context, horizontal, vertical int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.ScrollMouse(horizontal, vertical)
}
