package gopilot

import (
	"context"

	cdppage "github.com/mafredri/cdp/protocol/page"
)

// PageTakeScreenshotInput specifies input parameters for taking a screenshot of a page.
type PageTakeScreenshotInput struct {
	// Format specifies the desired image format for the screenshot.
	// Options could be "png" or "jpeg".
	Format string

	// Full determines whether to capture the entire page or only the current viewport.
	Full bool

	// Viewport allows specifying a custom area of the page to capture.
	Viewport *cdppage.Viewport
}

// PageTakeScreenshotOutput represents the output of the TakeScreenshot method for a page.
type PageTakeScreenshotOutput struct {
	// Data contains the base64 encoded screenshot image data.
	Data []byte
}

// TakeScreenshot captures a screenshot of the page.
// You can choose to capture the entire page or just the visible viewport.
// Input parameters allow you to specify the image format and capture area.
// Returns the base64 encoded screenshot data or an error if the capture fails.
func (p *page) TakeScreenshot(ctx context.Context, in *PageTakeScreenshotInput) (*PageTakeScreenshotOutput, error) {
	format := in.Format
	if format == "" {
		format = "png"
	}

	rp, err := p.client.Page.CaptureScreenshot(ctx, &cdppage.CaptureScreenshotArgs{
		Format:                &format,
		CaptureBeyondViewport: &in.Full,
		Clip:                  in.Viewport,
	})

	if err != nil {
		return nil, err
	}

	return &PageTakeScreenshotOutput{Data: rp.Data}, nil
}
