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
	node      dom.Node             // The DOM node representing the element.
	remoteObj runtime.RemoteObject // javascript object of the node
	client    *cdp.Client          // The CDP client for communication with the Chromium instance.
}

// newElement creates a new Element instance.
// It takes a DOM node, DevTools instance, and CDP client as parameters.
// Returns a new Element implementation.
func newElement(node dom.Node, remoteObj runtime.RemoteObject, client *cdp.Client) Element {
	return &element{
		node:      node,
		remoteObj: remoteObj,
		client:    client,
	}
}

func (e *element) GetNodeID(_ context.Context) dom.NodeID {
	return e.node.NodeID
}
