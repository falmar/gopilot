package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/dom"
)

// ElementScrollIntoViewInput contains parameters for the ScrollIntoView action.
type ElementScrollIntoViewInput struct{}

// ElementScrollIntoViewOutput contains the result of the ScrollIntoView action.
type ElementScrollIntoViewOutput struct {
	// X is the X coordinate of the scroll-to view position.
	X float64 `json:"x"`

	// Y is the Y coordinate of the scroll-to view position.
	Y float64 `json:"y"`
}

// ScrollIntoView executes a scroll-to-view action on the element.
// It ensures the element is within the viewport.
// Returns an ElementScrollIntoViewOutput or an error if the action fails.
func (e *element) ScrollIntoView(ctx context.Context, in *ElementScrollIntoViewInput) (*ElementScrollIntoViewOutput, error) {
	// Attempt to scroll the element into the view if needed.
	err := e.client.DOM.ScrollIntoViewIfNeeded(ctx, &dom.ScrollIntoViewIfNeededArgs{
		BackendNodeID: &e.node.BackendNodeID,
	})
	if err != nil {
		return nil, err
	}

	return &ElementScrollIntoViewOutput{}, nil
}
