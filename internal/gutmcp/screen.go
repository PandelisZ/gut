package gutmcp

import (
	"context"
	"fmt"

	gutpkg "github.com/PandelisZ/gut"
	"github.com/PandelisZ/gut/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerScreenTools(server *mcp.Server) {
	mcp.AddTool(server, readOnlyTool(
		"screen_capture",
		"Screen Capture",
		"Capture the full screen or a logical region and return image content plus geometry metadata.",
	), s.screenCaptureTool)

	mcp.AddTool(server, readOnlyTool(
		"screen_color_at",
		"Screen Color At",
		"Read the color at a logical screen coordinate.",
	), s.screenColorAtTool)

	mcp.AddTool(server, readOnlyTool(
		"screen_find_color",
		"Screen Find Color",
		"Find exact color matches on the screen or within an optional search region.",
	), s.screenFindColorTool)
}

func (s *Service) screenCaptureTool(ctx context.Context, _ *mcp.CallToolRequest, input ScreenCaptureInput) (*mcp.CallToolResult, ScreenCaptureOutput, error) {
	image, region, err := s.captureImage(ctx, input.Region)
	if err != nil {
		return nil, ScreenCaptureOutput{}, actionableError("screen_capture", err)
	}

	data, err := encodePNG(image)
	if err != nil {
		return nil, ScreenCaptureOutput{}, actionableError("screen_capture", err)
	}

	density := image.NormalizedPixelDensity()
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.ImageContent{
				MIMEType: "image/png",
				Data:     data,
			},
		},
	}
	return result, ScreenCaptureOutput{
		Format:         "image/png",
		CapturedRegion: regionToJSON(region),
		PixelSize: JSONSize{
			Width:  image.Width,
			Height: image.Height,
		},
		PixelDensity: JSONPixelDensity{
			ScaleX: density.ScaleX,
			ScaleY: density.ScaleY,
		},
	}, nil
}

func (s *Service) screenColorAtTool(ctx context.Context, _ *mcp.CallToolRequest, input ScreenColorAtInput) (*mcp.CallToolResult, ScreenColorAtOutput, error) {
	color, err := s.nut.Screen.ColorAt(ctx, input.Point.toShared())
	if err != nil {
		return nil, ScreenColorAtOutput{}, actionableError("screen_color_at", err)
	}
	return nil, ScreenColorAtOutput{
		Point: pointToJSON(input.Point.toShared()),
		Color: colorToJSON(color),
	}, nil
}

func (s *Service) screenFindColorTool(ctx context.Context, _ *mcp.CallToolRequest, input ScreenFindColorInput) (*mcp.CallToolResult, ScreenFindColorOutput, error) {
	color, err := input.Color.toShared()
	if err != nil {
		return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", err)
	}
	if input.Limit < 0 {
		return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", fmt.Errorf("limit must be non-negative"))
	}

	var searchRegion *shared.Region
	var searchRegionJSON *JSONRegion
	if input.SearchRegion != nil {
		if err := input.SearchRegion.validate(); err != nil {
			return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", err)
		}
		if input.SearchRegion.Width == 0 || input.SearchRegion.Height == 0 {
			return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", fmt.Errorf("searchRegion width and height must be greater than zero"))
		}
		region := input.SearchRegion.toShared()
		searchRegion = &region
		regionJSON := regionToJSON(region)
		searchRegionJSON = &regionJSON
	}

	results, err := s.nut.Screen.FindAll(ctx, gutpkg.PixelWithColor(color), &gutpkg.ScreenFindOptions{
		SearchRegion: searchRegion,
	})
	if err != nil {
		return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", err)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	points := make([]JSONPoint, 0, min(limit, len(results)))
	for idx, result := range results {
		if idx >= limit {
			break
		}
		point, ok := result.(shared.Point)
		if !ok {
			return nil, ScreenFindColorOutput{}, actionableError("screen_find_color", fmt.Errorf("unexpected screen_find_color result type %T", result))
		}
		points = append(points, pointToJSON(point))
	}

	return nil, ScreenFindColorOutput{
		Color:        colorToJSON(color),
		SearchRegion: searchRegionJSON,
		TotalMatches: len(results),
		Matches:      points,
	}, nil
}

func (s *Service) captureImage(ctx context.Context, regionInput *RegionInput) (shared.Image, shared.Region, error) {
	screenProvider, err := s.registry.Screen()
	if err != nil {
		return shared.Image{}, shared.Region{}, err
	}

	if regionInput == nil {
		region, err := screenProvider.ScreenSize(ctx)
		if err != nil {
			return shared.Image{}, shared.Region{}, err
		}
		image, err := s.nut.Screen.Grab(ctx)
		return image, region, err
	}

	if err := regionInput.validate(); err != nil {
		return shared.Image{}, shared.Region{}, err
	}
	if regionInput.Width == 0 || regionInput.Height == 0 {
		return shared.Image{}, shared.Region{}, fmt.Errorf("capture region width and height must be greater than zero")
	}
	region := regionInput.toShared()
	image, err := s.nut.Screen.GrabRegion(ctx, region)
	return image, region, err
}
