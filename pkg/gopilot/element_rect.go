package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/dom"
)

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
