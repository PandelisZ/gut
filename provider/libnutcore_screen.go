package provider

import (
	"context"
	"fmt"
	"time"

	"gut/native/common"
	"gut/shared"
)

type libnutcoreScreenProvider struct {
	client libnutcoreClient
}

func NewLibnutcoreScreenProvider(client libnutcoreClient) ScreenProvider {
	return &libnutcoreScreenProvider{client: client}
}

func (p *libnutcoreScreenProvider) GrabScreen(ctx context.Context) (shared.Image, error) {
	if err := ctx.Err(); err != nil {
		return shared.Image{}, err
	}
	bitmap, err := p.client.CaptureScreen(nil)
	if err != nil {
		return shared.Image{}, err
	}
	size, err := p.client.GetScreenSize()
	if err != nil {
		return shared.Image{}, err
	}
	return imageFromBitmap(bitmap, shared.Region{Left: 0, Top: 0, Width: size.Width, Height: size.Height}, "grabScreenResult")
}

func (p *libnutcoreScreenProvider) GrabScreenRegion(ctx context.Context, region shared.Region) (shared.Image, error) {
	if err := ctx.Err(); err != nil {
		return shared.Image{}, err
	}
	nativeRegion := regionToNative(region)
	bitmap, err := p.client.CaptureScreen(&nativeRegion)
	if err != nil {
		return shared.Image{}, err
	}
	return imageFromBitmap(bitmap, region, "grabScreenRegionResult")
}

func (p *libnutcoreScreenProvider) HighlightScreenRegion(ctx context.Context, region shared.Region, duration time.Duration, opacity float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.Highlight(regionToNative(region), duration, opacity)
}

func (p *libnutcoreScreenProvider) ScreenWidth(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	size, err := p.client.GetScreenSize()
	if err != nil {
		return 0, err
	}
	return size.Width, nil
}

func (p *libnutcoreScreenProvider) ScreenHeight(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	size, err := p.client.GetScreenSize()
	if err != nil {
		return 0, err
	}
	return size.Height, nil
}

func (p *libnutcoreScreenProvider) ScreenSize(ctx context.Context) (shared.Region, error) {
	if err := ctx.Err(); err != nil {
		return shared.Region{}, err
	}
	size, err := p.client.GetScreenSize()
	if err != nil {
		return shared.Region{}, err
	}
	return shared.Region{Left: 0, Top: 0, Width: size.Width, Height: size.Height}, nil
}

func imageFromBitmap(bitmap *common.Bitmap, region shared.Region, id string) (shared.Image, error) {
	if bitmap == nil {
		return shared.Image{}, fmt.Errorf("bitmap is nil")
	}
	scaleX := 1.0
	scaleY := 1.0
	if region.Width > 0 {
		scaleX = float64(bitmap.Width) / float64(region.Width)
	}
	if region.Height > 0 {
		scaleY = float64(bitmap.Height) / float64(region.Height)
	}
	return shared.NewImage(
		bitmap.Width,
		bitmap.Height,
		bitmap.Image,
		bitmap.BytesPerPixel,
		id,
		bitmap.BitsPerPixel,
		bitmap.ByteWidth,
		shared.ColorModeBGR,
		shared.PixelDensity{ScaleX: scaleX, ScaleY: scaleY},
	)
}
