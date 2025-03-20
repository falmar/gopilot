package gopilot

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
)

// Element represents an interactive element in a web page.
type Element interface {
	// Click simulates a mouse click on the element.
	// Accepts an ElementClickInput containing details for the click action.
	// Returns an ElementClickOutput with the result or an error if the click fails.
	Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error)

	// GetRect retrieves the bounding rectangle of the element.
	// Returns a BoundingRect containing the dimensions and position of the element or an error if retrieval fails.
	GetRect(ctx context.Context) (*BoundingRect, error)
}

// element is an implementation of the Element interface.
type element struct {
	node   dom.Node          // The DOM node representing the element.
	tools  *devtool.DevTools // The DevTools instance for interacting with the browser.
	client *cdp.Client       // The CDP client for communication with the Chromium instance.
}

// newElement creates a new Element instance.
// It takes a DOM node, DevTools instance, and CDP client as parameters.
// Returns a new Element implementation.
func newElement(node dom.Node, tools *devtool.DevTools, client *cdp.Client) Element {
	return &element{
		node:   node,
		tools:  tools,
		client: client,
	}
}
