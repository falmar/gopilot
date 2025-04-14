package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp/protocol/input"
)

type PageInput interface {
	// TypeText sends a sequence of keystrokes to the element as if typed by a user.
	// Accepts an ElementTypeInput containing the text to type.
	// Returns an ElementTypeOutput with the result or an error if typing fails.
	TypeText(ctx context.Context, in *PageTypeTextInput) (*PageTypeTextOutput, error)
}

type DispatchEventType string

const (
	DispatchEventTypeKeyDown    DispatchEventType = "keyDown"
	DispatchEventTypeKeyUp      DispatchEventType = "keyUp"
	DispatchEventTypeRawKeyDown DispatchEventType = "rawKeyDown"
	DispatchEventTypeChar       DispatchEventType = "char"
)

type TypeDelayFunc func() time.Duration

// PageTypeTextInput specifies input for the Type method.
// Text specifies the string to type into the page.
// Delay is the duration between keystrokes.
// DelayFunc is function to control typing delays.
type PageTypeTextInput struct {
	Text      string        // The text to be typed into the page.
	Delay     time.Duration // (optional) Duration between keystrokes.
	DelayFunc TypeDelayFunc // (optional) Custom function for typing delays.
}

// PageTypeTextOutput represents the output of the Type method.
// It is currently empty, but can be extended to provide additional details of the typing operation.
type PageTypeTextOutput struct{}

// TypeText sends a sequence of keystrokes to the page as if typed by a user.
// Accepts an PageTypeInput containing the text to type.
// Returns an PageTypeOutput with the result or an error if typing fails.
func (p *page) TypeText(ctx context.Context, in *PageTypeTextInput) (*PageTypeTextOutput, error) {
	last := len(in.Text)
	for i, t := range in.Text {
		t := string(t)
		var toType *string = nil
		var key *string = nil
		var code *string = nil
		var nativeVirtualCode *int = nil
		var keyIdentifier *string = nil

		keyDown := DispatchEventTypeKeyDown

		if t == " " || t == "\u00A0" {
			t = " "
			keyDown = DispatchEventTypeRawKeyDown
			k := " "
			c := "Space"
			vc := 32
			ki := "U+0020"

			key = &k
			code = &c
			nativeVirtualCode = &vc
			keyIdentifier = &ki
			toType = &t
		} else {
			toType = &t
		}

		err := p.client.Input.DispatchKeyEvent(ctx, &input.DispatchKeyEventArgs{
			Type:                  string(keyDown),
			Text:                  toType,
			UnmodifiedText:        toType,
			Key:                   key,
			Code:                  code,
			NativeVirtualKeyCode:  nativeVirtualCode,
			WindowsVirtualKeyCode: nativeVirtualCode,
			KeyIdentifier:         keyIdentifier,
		})
		if err != nil {
			return nil, err
		}

		err = p.client.Input.DispatchKeyEvent(ctx, &input.DispatchKeyEventArgs{
			Type:                  string(DispatchEventTypeKeyUp),
			Key:                   key,
			Code:                  code,
			NativeVirtualKeyCode:  nativeVirtualCode,
			WindowsVirtualKeyCode: nativeVirtualCode,
			KeyIdentifier:         keyIdentifier,
		})
		if err != nil {
			return nil, err
		}

		if i == last-1 {
			break
		}

		if in.Delay > 0 {
			if err = sleepWithCtx(ctx, in.Delay); err != nil {
				return nil, err
			}
		} else if in.DelayFunc != nil {
			if err = sleepWithCtx(ctx, in.DelayFunc()); err != nil {
				return nil, err
			}
		}
	}

	return &PageTypeTextOutput{}, nil
}
