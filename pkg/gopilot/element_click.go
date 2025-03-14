package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp/protocol/input"
)

// ElementClickInput
//
// StepDuration set to sleep for a duration on each step:
// 1) move to element
// 2) mousePress
// 3) mouseRelease
//
// HoldDuration duration to wait between mousePress and mouseRelease.
// if not set, defaults to StepDuration
type ElementClickInput struct {
	StepDuration time.Duration
	HoldDuration time.Duration
}

// ElementClickOutput returns the position clicked
type ElementClickOutput struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (e *element) Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error) {
	rect, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type: "mouseMoved",
		X:    rect.X,
		Y:    rect.Y,
	})
	if err != nil {
		return nil, err
	}

	if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	clientCount := 1

	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mousePressed",
		Button:     input.MouseButtonLeft,
		X:          rect.X,
		Y:          rect.Y,
		ClickCount: &clientCount,
	})
	if err != nil {
		return nil, err
	}

	if in.HoldDuration > 0 {
		time.Sleep(in.HoldDuration)
	} else if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mouseReleased",
		Button:     input.MouseButtonLeft,
		X:          rect.X,
		Y:          rect.Y,
		ClickCount: &clientCount,
	})
	if err != nil {
		return nil, err
	}

	return &ElementClickOutput{X: rect.X, Y: rect.Y}, nil
}
