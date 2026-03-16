package provider

import (
	"context"
	"time"

	"gut/native/common"
	"gut/shared"
)

type libnutcoreKeyboardProvider struct {
	client libnutcoreClient
}

func NewLibnutcoreKeyboardProvider(client libnutcoreClient) KeyboardProvider {
	return &libnutcoreKeyboardProvider{client: client}
}

func (p *libnutcoreKeyboardProvider) SetKeyboardDelay(delay time.Duration) {
	_ = p.client.SetKeyboardDelay(delay)
}

func (p *libnutcoreKeyboardProvider) Type(ctx context.Context, input string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.TypeString(input)
}

func (p *libnutcoreKeyboardProvider) Click(ctx context.Context, keys ...shared.Key) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	if err := p.PressKey(ctx, keys...); err != nil {
		return err
	}
	return p.ReleaseKey(ctx, keys...)
}

func (p *libnutcoreKeyboardProvider) PressKey(ctx context.Context, keys ...shared.Key) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, key := range keys {
		token, err := keyToLibnutToken(key)
		if err != nil {
			return err
		}
		if err := p.client.KeyToggle(token, common.KeyStateDown); err != nil {
			return err
		}
	}
	return nil
}

func (p *libnutcoreKeyboardProvider) ReleaseKey(ctx context.Context, keys ...shared.Key) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for i := len(keys) - 1; i >= 0; i-- {
		token, err := keyToLibnutToken(keys[i])
		if err != nil {
			return err
		}
		if err := p.client.KeyToggle(token, common.KeyStateUp); err != nil {
			return err
		}
	}
	return nil
}
