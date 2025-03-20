package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp/protocol/input"
)

// ElementClickInput specifies the input parameters for simulating a click on an element.
// StepDuration sets the duration to wait between each step of the click process:
// 1) move to the element
// 2) mouse press
// 3) mouse release
// HoldDuration specifies the duration to wait between mousePress and mouseRelease.
// If not set, it defaults to StepDuration.
type ElementClickInput struct {
	StepDuration time.Duration // Duration for each step of the click action.
	HoldDuration time.Duration // Duration to hold the mouse press before releasing.
}

// ElementClickOutput represents the output of a click action.
// It returns the X and Y coordinates where the click occurred.
type ElementClickOutput struct {
	X float64 `json:"x"` // X coordinate of the click position.
	Y float64 `json:"y"` // Y coordinate of the click position.
}

// Click simulates a click on the element.
// It receives a context and an ElementClickInput as parameters.
// Returns an ElementClickOutput with the click coordinates or an error if the click action fails.
func (e *element) Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error) {
	rect, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	// Move the mouse to the element's position.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type: "mouseMoved",
		X:    rect.X,
		Y:    rect.Y,
	})
	if err != nil {
		return nil, err
	}

	// Wait for StepDuration before pressing the mouse button.
	if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	clientCount := 1

	// Press the mouse button.
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

	// Wait for HoldDuration or StepDuration before releasing the mouse button.
	if in.HoldDuration > 0 {
		time.Sleep(in.HoldDuration)
	} else if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	// Release the mouse button.
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

	// Return the coordinates of the click.
	return &ElementClickOutput{X: rect.X, Y: rect.Y}, nil
}
