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
	primary, modifiers, err := splitPrimaryAndModifiers(keys)
	if err != nil {
		return err
	}
	return p.client.KeyTap(primary, modifiers...)
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

func splitPrimaryAndModifiers(keys []shared.Key) (string, []string, error) {
	primary, err := keyToLibnutToken(keys[len(keys)-1])
	if err != nil {
		return "", nil, err
	}
	modifiers := make([]string, 0, len(keys)-1)
	for _, key := range keys[:len(keys)-1] {
		token, ok := modifierToLibnutToken(key)
		if ok {
			modifiers = append(modifiers, token)
			continue
		}
		token, err = keyToLibnutToken(key)
		if err != nil {
			return "", nil, err
		}
		modifiers = append(modifiers, token)
	}
	return primary, modifiers, nil
}
