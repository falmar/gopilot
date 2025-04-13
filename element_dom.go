package gopilot

import (
	"context"
	"encoding/json"

	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/runtime"
)

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

func (e *element) Remove(ctx context.Context) error {
	return e.client.DOM.RemoveNode(ctx, &dom.RemoveNodeArgs{
		NodeID: e.node.NodeID,
	})
}
