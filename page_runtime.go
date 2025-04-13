package gopilot

import (
	"context"
	"encoding/json"

	"github.com/mafredri/cdp/protocol/runtime"
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

	res, err := p.client.Runtime.Evaluate(ctx, &runtime.EvaluateArgs{
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
