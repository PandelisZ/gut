package provider

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"gut/native/common"
	"gut/shared"
)

var libnutcoreScreenGOOS = runtime.GOOS

var libnutcoreMacOSScreencapture = runMacOSScreencapture

func captureMacOSScreenFallback(ctx context.Context, nativeRegion *common.Rect, region shared.Region, id string) (shared.Image, error) {
	data, err := libnutcoreMacOSScreencapture(ctx, nativeRegion)
	if err != nil {
		return shared.Image{}, err
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return shared.Image{}, err
	}
	return imageFromDecoded(decoded, region, id)
}

func imageFromDecoded(decoded image.Image, region shared.Region, id string) (shared.Image, error) {
	if decoded == nil {
		return shared.Image{}, fmt.Errorf("decoded image is nil")
	}
	bounds := decoded.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(rgba, rgba.Bounds(), decoded, bounds.Min, draw.Src)
	scaleX := 1.0
	scaleY := 1.0
	if region.Width > 0 {
		scaleX = float64(rgba.Bounds().Dx()) / float64(region.Width)
	}
	if region.Height > 0 {
		scaleY = float64(rgba.Bounds().Dy()) / float64(region.Height)
	}
	return shared.NewImage(
		rgba.Bounds().Dx(),
		rgba.Bounds().Dy(),
		rgba.Pix,
		4,
		id,
		32,
		rgba.Stride,
		shared.ColorModeRGB,
		shared.PixelDensity{ScaleX: scaleX, ScaleY: scaleY},
	)
}

func libnutcoreCapabilityUnavailableError(client libnutcoreClient, operation string, capability common.Capability) error {
	status := client.Capabilities().Status(capability)
	info := client.Info()
	platform := info.Platform
	if platform == "" {
		platform = libnutcoreScreenGOOS
	}
	details := make([]string, 0, 3)
	if status.Reason != "" {
		details = append(details, status.Reason)
	}
	if info.Name != "" {
		details = append(details, fmt.Sprintf("backend=%s", info.Name))
	}
	if info.BindingState != "" {
		details = append(details, fmt.Sprintf("binding=%s", info.BindingState))
	}
	if len(info.Notes) > 0 {
		details = append(details, strings.Join(info.Notes, "; "))
	}
	detail := strings.Join(details, "; ")
	if detail == "" {
		detail = "capability unavailable"
	}
	return common.CapabilityUnavailable(operation, platform, capability, detail)
}

func runMacOSScreencapture(ctx context.Context, region *common.Rect) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "gut-provider-screencapture-*.png")
	if err != nil {
		return nil, err
	}
	path := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	defer os.Remove(path)
	args := []string{"-x", "-t", "png"}
	if region != nil {
		args = append(args, fmt.Sprintf("-R%d,%d,%d,%d", region.X, region.Y, region.Width, region.Height))
	}
	args = append(args, path)
	output, err := exec.CommandContext(ctx, "screencapture", args...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return nil, fmt.Errorf("screencapture failed: %s: %w", message, err)
		}
		return nil, fmt.Errorf("screencapture failed: %w", err)
	}
	return os.ReadFile(path)
}
