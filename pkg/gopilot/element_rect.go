package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/dom"
)

type BoundingRect struct {
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Bottom float64 `json:"bottom"`
	Right  float64 `json:"right"`

	X float64 `json:"x"`
	Y float64 `json:"y"`

	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

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

	return &BoundingRect{
		Left:  quad[0],
		Top:   quad[1],
		Right: quad[2],

		// top [3]
		// right [4]

		Bottom: quad[5],

		X: quad[0],
		Y: quad[1],

		Width:  float64(brp.Model.Width),
		Height: float64(brp.Model.Height),
	}, nil
}
