package gopilot

import (
	"context"
	"encoding/json"

	cdppage "github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

// PageEvaluateInput specifies input for the Evaluate method.
type PageEvaluateInput struct {
	AwaitPromise bool
	ReturnValue  bool
	Expression   string
}

// PageEvaluateOutput represents the output of the Evaluate method.
type PageEvaluateOutput struct {
	Value json.RawMessage
}

// Evaluate executes the given JavaScript expression on the page.
func (p *page) Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error) {
	userGesture := true
	allowUnsafe := true

	res, err := p.client.Runtime.Evaluate(ctx, &cdpruntime.EvaluateArgs{
		Expression:                  in.Expression,
		UserGesture:                 &userGesture,
		ReturnByValue:               &in.ReturnValue,
		AwaitPromise:                &in.AwaitPromise,
		AllowUnsafeEvalBlockedByCSP: &allowUnsafe,
	})
	if err != nil {
		return nil, err
	}

	out := &PageEvaluateOutput{}
	if in.ReturnValue {
		out.Value = res.Result.Value
	}

	return out, nil
}

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
	// Data contains the screenshot image data.
	Data []byte
}

// TakeScreenshot captures a screenshot of the page.
// You can choose to capture the entire page or just the visible viewport.
// Input parameters allow you to specify the image format and capture area.
// Returns the encoded screenshot data or an error if the capture fails.
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
