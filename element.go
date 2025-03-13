package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/input"
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

type Element interface {
	Click(ctx context.Context, in *ElementClientInput) (*ElementClientOutput, error)

	GetRect(ctx context.Context) (BoundingRect, error)
}

type element struct {
	node dom.Node

	tools  *devtool.DevTools
	client *cdp.Client

	boxModel dom.BoxModel
}

func newElement(
	node dom.Node,
	tools *devtool.DevTools,
	client *cdp.Client,
) Element {
	return &element{
		node:   node,
		tools:  tools,
		client: client,
	}
}

// ElementClientInput
//
// StepDuration set to sleep for a duration on each step:
// 1) move to element
// 2) mousePress
// 3) mouseRelease
//
// HoldDuration duration to wait between mousePress and mouseRelease.
// if not set, defaults to StepDuration
type ElementClientInput struct {
	StepDuration time.Duration
	HoldDuration time.Duration
}

// ElementClientOutput returns the position clicked
type ElementClientOutput struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (e *element) Click(ctx context.Context, in *ElementClientInput) (*ElementClientOutput, error) {
	rect, err := e.GetRect(ctx)
	if err != nil {
		return nil, err
	}

	err = e.client.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type: "mouseMoved",
		X:    rect.X,
		Y:    rect.Y,
	})
	if err != nil {
		return nil, err
	}

	if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

	clientCount := 1

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

	if in.HoldDuration > 0 {
		time.Sleep(in.HoldDuration)
	} else if in.StepDuration > 0 {
		time.Sleep(in.StepDuration)
	}

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

	return &ElementClientOutput{X: rect.X, Y: rect.Y}, nil
}

func (e *element) GetRect(ctx context.Context) (BoundingRect, error) {
	qrp, err := e.client.DOM.GetContentQuads(ctx, &dom.GetContentQuadsArgs{
		BackendNodeID: &e.node.BackendNodeID,
	})
	if err != nil {
		return BoundingRect{}, err
	}

	brp, err := e.client.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
		BackendNodeID: &e.node.BackendNodeID,
	})
	if err != nil {
		return BoundingRect{}, err
	}

	quad := qrp.Quads[0]

	return BoundingRect{
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
