package provider

import (
	"context"

	"github.com/PandelisZ/gut/shared"
)

type libnutcoreWindowProvider struct {
	client libnutcoreClient
}

func NewLibnutcoreWindowProvider(client libnutcoreClient) WindowProvider {
	return &libnutcoreWindowProvider{client: client}
}

func (p *libnutcoreWindowProvider) GetWindows(ctx context.Context) ([]shared.WindowHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	handles, err := p.client.GetWindows()
	if err != nil {
		return nil, err
	}
	windows := make([]shared.WindowHandle, 0, len(handles))
	for _, handle := range handles {
		windows = append(windows, windowHandleFromNative(handle))
	}
	return windows, nil
}

func (p *libnutcoreWindowProvider) GetActiveWindow(ctx context.Context) (shared.WindowHandle, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	handle, err := p.client.GetActiveWindow()
	if err != nil {
		return 0, err
	}
	return windowHandleFromNative(handle), nil
}

func (p *libnutcoreWindowProvider) GetWindowTitle(ctx context.Context, windowHandle shared.WindowHandle) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return p.client.GetWindowTitle(windowHandleToNative(windowHandle))
}

func (p *libnutcoreWindowProvider) GetWindowRegion(ctx context.Context, windowHandle shared.WindowHandle) (shared.Region, error) {
	if err := ctx.Err(); err != nil {
		return shared.Region{}, err
	}
	region, err := p.client.GetWindowRect(windowHandleToNative(windowHandle))
	if err != nil {
		return shared.Region{}, err
	}
	return regionFromNative(region), nil
}

func (p *libnutcoreWindowProvider) FocusWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.FocusWindow(windowHandleToNative(windowHandle))
}

func (p *libnutcoreWindowProvider) MoveWindow(ctx context.Context, windowHandle shared.WindowHandle, newOrigin shared.Point) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.MoveWindow(windowHandleToNative(windowHandle), pointToNative(newOrigin))
}

func (p *libnutcoreWindowProvider) ResizeWindow(ctx context.Context, windowHandle shared.WindowHandle, newSize shared.Size) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.ResizeWindow(windowHandleToNative(windowHandle), sizeToNative(newSize))
}

func (p *libnutcoreWindowProvider) MinimizeWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.MinimizeWindow(windowHandleToNative(windowHandle))
}

func (p *libnutcoreWindowProvider) RestoreWindow(ctx context.Context, windowHandle shared.WindowHandle) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return p.client.RestoreWindow(windowHandleToNative(windowHandle))
}
