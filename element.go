package gopilot

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/runtime"
)

// Element represents an interactive element in a web page.
type Element interface {
	ElementInput
	ElementDOM

	// TakeScreenshot captures a screenshot of the element.
	// It uses the element's position and size to define the capture area.
	// Input parameters can specify the format of the image.
	// Returns the screenshot data as bytes or an error if the capture fails.
	TakeScreenshot(ctx context.Context, in *ElementTakeScreenshotInput) (*ElementTakeScreenshotOutput, error)

	// GetNodeID gives the current node of the element
	GetNodeID(ctx context.Context) dom.NodeID
}

// element is an implementation of the Element interface.
type element struct {
	node   dom.Node    // The DOM node representing the element.
	client *cdp.Client // The CDP client for communication with the Chromium instance.
}

// NewElement creates a new Element instance.
// It takes a DOM node, DevTools instance, and CDP client as parameters.
// Returns a new Element implementation.
func NewElement(client *cdp.Client, node dom.Node) Element {
	return &element{
		client: client,
		node:   node,
	}
}

func (e *element) getRemoteObject(ctx context.Context) (*runtime.RemoteObject, error) {
	rrp, err := e.client.DOM.ResolveNode(ctx, &dom.ResolveNodeArgs{
		NodeID: &e.node.NodeID,
	})
	if err != nil {
		return nil, err
	}

	return &rrp.Object, nil
}

func (e *element) GetNodeID(_ context.Context) dom.NodeID {
	return e.node.NodeID
}
