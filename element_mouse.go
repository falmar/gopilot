package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp/protocol/input"
)

// ElementClickInput specifies the input parameters for simulating a click on an element.
// - StepDuration: Duration to wait between each step of the click process: moving to the element, mouse press, and mouse release.
// - HoldDuration: Duration to wait between mouse press and mouse release. Defaults to StepDuration if not set.
type ElementClickInput struct {
	StepDuration time.Duration // Duration for each step of the click action.
	HoldDuration time.Duration // Duration to hold the mouse press before releasing.

	ReturnHoldRelease bool // Return a release function to let user decide when to release mouse press
}

// ElementClickOutput represents the output of a click action.
// It provides the X and Y coordinates where the click occurred.
type ElementClickOutput struct {
	X float64 `json:"x"` // X coordinate of the click position.
	Y float64 `json:"y"` // Y coordinate of the click position.

	Release func() error
}

// Click simulates a mouse click on the element.
// It calculates the center of the element and executes a mouse click at the center.
// You can use
func (e *element) Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error) {
	rect, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	// Move the mouse to the center of the element.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type: "mouseMoved",
		X:    rect.CenterX,
		Y:    rect.CenterY,
	})
	if err != nil {
		return nil, err
	}

	// Wait for StepDuration before pressing the mouse button.
	if in.StepDuration > 0 {
		if err = sleepWithCtx(ctx, in.StepDuration); err != nil {
			return nil, err
		}
	}

	clientCount := 1

	// Press the mouse button at the center of the element.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mousePressed",
		Button:     input.MouseButtonLeft,
		X:          rect.CenterX,
		Y:          rect.CenterY,
		ClickCount: &clientCount,
	})
	if err != nil {
		return nil, err
	}

	// Release the mouse button at the center of the element.
	release := func() error {
		return e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
			Type:       "mouseReleased",
			Button:     input.MouseButtonLeft,
			X:          rect.CenterX,
			Y:          rect.CenterY,
			ClickCount: &clientCount,
		})
	}

	if in.ReturnHoldRelease {
		return &ElementClickOutput{X: rect.CenterX, Y: rect.CenterY, Release: release}, nil
	}

	// Wait for HoldDuration or default to StepDuration before releasing the mouse button.
	if in.HoldDuration > 0 {
		if err = sleepWithCtx(ctx, in.HoldDuration); err != nil {
			return nil, err
		}
	} else if in.StepDuration > 0 {
		if err = sleepWithCtx(ctx, in.StepDuration); err != nil {
			return nil, err
		}
	}

	if err = release(); err != nil {
		return nil, err
	}

	// Return the coordinates where the click was performed.
	return &ElementClickOutput{X: rect.CenterX, Y: rect.CenterY}, nil
}
