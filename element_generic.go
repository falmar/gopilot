package gopilot

import (
	"context"

	cdppage "github.com/mafredri/cdp/protocol/page"
)

// ElementTakeScreenshotInput specifies input parameters for taking a screenshot of an element.
type ElementTakeScreenshotInput struct {
	// Format specifies the desired image format for the screenshot.
	// Common formats include "png" and "jpeg".
	Format string
}

// ElementTakeScreenshotOutput represents the output of the TakeScreenshot method for an element.
type ElementTakeScreenshotOutput struct {
	// Data contains the base64 encoded screenshot image data.
	Data []byte
}

// TakeScreenshot captures a screenshot of the element.
// It uses the element's position and size to define the capture area.
// Input parameters can specify the format of the image.
// Returns the screenshot data as base64 encoded bytes or an error if the capture fails.
func (e *element) TakeScreenshot(ctx context.Context, in *ElementTakeScreenshotInput) (*ElementTakeScreenshotOutput, error) {
	rrp, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	format := in.Format
	if format == "" {
		format = "png"
	}

	vp := true
	var scale float64 = 1

	rp, err := e.client.Page.CaptureScreenshot(ctx, &cdppage.CaptureScreenshotArgs{
		Format:                &format,
		CaptureBeyondViewport: &vp,
		Clip: &cdppage.Viewport{
			X: rrp.X, Y: rrp.Y, Scale: scale,
			Width: rrp.Width, Height: rrp.Height,
		},
	})

	if err != nil {
		return nil, err
	}

	return &ElementTakeScreenshotOutput{Data: rp.Data}, nil
}
