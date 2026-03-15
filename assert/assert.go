package assert

import (
	"context"
	"fmt"

	"gut/screen"
	"gut/shared"
)

type Assert struct {
	screen *screen.Screen
}

func New(scr *screen.Screen) *Assert {
	return &Assert{screen: scr}
}

func (a *Assert) Visible(ctx context.Context, input any, searchRegion *shared.Region, confidence *float64) error {
	_, err := a.screen.Find(ctx, input, &screen.FindOptions{SearchRegion: searchRegion, Confidence: confidence})
	if err != nil {
		return fmt.Errorf("expected input to be visible: %w", err)
	}
	return nil
}

func (a *Assert) NotVisible(ctx context.Context, input any, searchRegion *shared.Region, confidence *float64) error {
	results, err := a.screen.FindAll(ctx, input, &screen.FindOptions{SearchRegion: searchRegion, Confidence: confidence})
	if err != nil {
		return err
	}
	if len(results) != 0 {
		return fmt.Errorf("expected input to not be visible")
	}
	return nil
}
