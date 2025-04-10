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
}

// ElementClickOutput represents the output of a click action.
// It provides the X and Y coordinates where the click occurred.
type ElementClickOutput struct {
	X float64 `json:"x"` // X coordinate of the click position.
	Y float64 `json:"y"` // Y coordinate of the click position.
}

// Click simulates a mouse click on the element.
// It calculates the center of the element and executes a mouse click at the center.
func (e *element) Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error) {
	rect, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate the center coordinates of the element for the click action.
	centerX := rect.X + rect.Width/2
	centerY := rect.Y + rect.Height/2

	// Move the mouse to the center of the element.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type: "mouseMoved",
		X:    centerX,
		Y:    centerY,
	})
	if err != nil {
		return nil, err
	}

	// Wait for StepDuration before pressing the mouse button.
	if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	clientCount := 1

	// Press the mouse button at the center of the element.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mousePressed",
		Button:     input.MouseButtonLeft,
		X:          centerX,
		Y:          centerY,
		ClickCount: &clientCount,
	})
	if err != nil {
		return nil, err
	}

	// Wait for HoldDuration or default to StepDuration before releasing the mouse button.
	if in.HoldDuration > 0 {
		time.Sleep(in.HoldDuration)
	} else if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	// Release the mouse button at the center of the element.
	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mouseReleased",
		Button:     input.MouseButtonLeft,
		X:          centerX,
		Y:          centerY,
		ClickCount: &clientCount,
	})
	if err != nil {
		return nil, err
	}

	// Return the coordinates where the click was performed.
	return &ElementClickOutput{X: centerX, Y: centerY}, nil
}
