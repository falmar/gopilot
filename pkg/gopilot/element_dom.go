package gopilot

import (
	"context"
	"encoding/json"

	"github.com/mafredri/cdp/protocol/runtime"
)

func (e *element) Text(ctx context.Context) (string, error) {
	returnByValue := true
	cfrp, err := e.client.Runtime.CallFunctionOn(ctx, &runtime.CallFunctionOnArgs{
		ObjectID:            e.remoteObj.ObjectID,
		ReturnByValue:       &returnByValue,
		FunctionDeclaration: `function() { return this.textContent; }`,
	})
	if err != nil {
		return "", err
	}

	var elementText string

	err = json.Unmarshal(cfrp.Result.Value, &elementText)
	if err != nil {
		return "", err
	}

	return elementText, nil
}
