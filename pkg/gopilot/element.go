package gopilot

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
)

type Element interface {
	Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error)
	GetRect(ctx context.Context) (*BoundingRect, error)
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
