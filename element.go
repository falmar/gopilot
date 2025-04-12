package gopilot

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/runtime"
)

// Element represents an interactive element in a web page.
type Element interface {
	// Click simulates a mouse click on the element.
	// Accepts an ElementClickInput containing details for the click action.
	// Returns an ElementClickOutput with the result or an error if the click fails.
	Click(ctx context.Context, in *ElementClickInput) (*ElementClickOutput, error)

	// ScrollIntoView performs an action to scroll the element into the viewport.
	// Accepts an ElementScrollIntoViewInput with scroll parameters.
	// Returns an ElementScrollIntoViewOutput or an error if the action fails.
	ScrollIntoView(ctx context.Context, in *ElementScrollIntoViewInput) (*ElementScrollIntoViewOutput, error)

	// Text retrieves the element's text content.
	Text(ctx context.Context) (string, error)

	// Focus sets focus on the element, allowing it to receive input.
	// Returns an error if the action fails.
	Focus(ctx context.Context) error

	// Remove the element from the DOM tree
	Remove(ctx context.Context) error

	// GetRect retrieves the bounding rectangle of the element.
	// Returns a BoundingRect containing the dimensions and position of the element, or an error if retrieval fails.
	GetRect(ctx context.Context) (*BoundingRect, error)
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
