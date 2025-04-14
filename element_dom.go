package gopilot

import (
	"context"
	"encoding/json"

	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/runtime"
)

type ElementDOM interface {
	// Text retrieves the element's text content.
	Text(ctx context.Context) (string, error)

	// Focus sets focus on the element, allowing it to receive input.
	// Returns an error if the action fails.
	Focus(ctx context.Context) error

	// ScrollIntoView performs an action to scroll the element into the viewport.
	// Accepts an ElementScrollIntoViewInput with scroll parameters.
	// Returns an ElementScrollIntoViewOutput or an error if the action fails.
	ScrollIntoView(ctx context.Context, in *ElementScrollIntoViewInput) (*ElementScrollIntoViewOutput, error)

	// GetRect retrieves the bounding rectangle of the element.
	// Returns a BoundingRect containing the dimensions and position of the element, or an error if retrieval fails.
	GetRect(ctx context.Context) (*BoundingRect, error)

	// Remove the element from the DOM tree
	Remove(ctx context.Context) error
}

func (e *element) Remove(ctx context.Context) error {
	return e.client.DOM.RemoveNode(ctx, &dom.RemoveNodeArgs{
		NodeID: e.node.NodeID,
	})
}

// Focus sets focus on the element, allowing it to receive input.
// Returns an error if the action fails.
func (e *element) Focus(ctx context.Context) error {
	return e.client.DOM.Focus(ctx, &dom.FocusArgs{
		ObjectID: e.remoteObj.ObjectID,
	})
}

func (e *element) Text(ctx context.Context) (string, error) {
	returnByValue := true
	rp, err := e.client.Runtime.CallFunctionOn(ctx, &runtime.CallFunctionOnArgs{
		ObjectID:            e.remoteObj.ObjectID,
		ReturnByValue:       &returnByValue,
		FunctionDeclaration: `function() { return this.textContent; }`,
	})
	if err != nil {
		return "", err
	}

	var elementText string

	err = json.Unmarshal(rp.Result.Value, &elementText)
	if err != nil {
		return "", err
	}

	return elementText, nil
}

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

// BoundingRect represents the bounding box of an element on the page.
// It contains the coordinates of the edges and dimensions of the element.
type BoundingRect struct {
	// Top is the distance from the top of the viewport to the top of the element.
	Top float64 `json:"top"`
	// Left is the distance from the left of the viewport to the left of the element.
	Left float64 `json:"left"`
	// Bottom is the distance from the top of the viewport to the bottom of the element.
	Bottom float64 `json:"bottom"`
	// Right is the distance from the left of the viewport to the right of the element.
	Right float64 `json:"right"`
	// X is the horizontal coordinate of the element.
	X float64 `json:"x"`
	// Y is the vertical coordinate of the element.
	Y float64 `json:"y"`
	// Width is the width of the element.
	Width float64 `json:"width"`
	// Height is the height of the element.
	Height float64 `json:"height"`
	// CenterX is the centered position on x-axis
	CenterX float64
	// CenterY is the centered position on y-axis
	CenterY float64
}

// GetRect retrieves the bounding rectangle of the element.
// It returns a BoundingRect containing the dimensions and position of the element,
// or an error if retrieving the rectangle fails.
func (e *element) GetRect(ctx context.Context) (*BoundingRect, error) {
	qrp, err := e.client.DOM.GetContentQuads(ctx, &dom.GetContentQuadsArgs{
		BackendNodeID: &e.node.BackendNodeID,
	})
	if err != nil {
		return nil, err
	}

	brp, err := e.client.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
		BackendNodeID: &e.node.BackendNodeID,
	})
	if err != nil {
		return nil, err
	}

	quad := qrp.Quads[0]

	rect := &BoundingRect{
		Left:   quad[0],
		Top:    quad[1],
		Right:  quad[2],
		Bottom: quad[5],
		X:      quad[0],
		Y:      quad[1],
		Width:  float64(brp.Model.Width),
		Height: float64(brp.Model.Height),
	}

	// Calculate the center coordinates of the element for the click action.
	rect.CenterX = rect.X + rect.Width/2
	rect.CenterY = rect.Y + rect.Height/2

	return rect, nil
}
