package gopilot

import (
	"context"
	"strings"
	"time"

	"github.com/mafredri/cdp/protocol/input"
)

type PageInput interface {
	// TypeText sends a sequence of keystrokes to the element as if typed by a user.
	// Accepts an ElementTypeInput containing the text to type.
	// Returns an ElementTypeOutput with the result or an error if typing fails.
	TypeText(ctx context.Context, in *PageTypeTextInput) (*PageTypeTextOutput, error)

	// PressKey dispatches key press events for special keys and key combinations.
	// Supports modifier keys (Ctrl, Alt, Shift, Meta) and repeat counts.
	PressKey(ctx context.Context, in *PagePressKeyInput) (*PagePressKeyOutput, error)
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

	// Deprecated: Use PressKey for special key handling instead.
	UseRawKeyDown bool
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
		var sysKey *bool = nil

		toType = &t

		if in.UseRawKeyDown && (t == " " || t == "\u00A0") {
			t = " "
			k := " "
			c := "Space"
			vc := 32
			ki := "U+0020"
			sk := true

			key = &k
			code = &c
			nativeVirtualCode = &vc
			keyIdentifier = &ki
			sysKey = &sk
			toType = &t

			err := p.client.Input.DispatchKeyEvent(ctx, &input.DispatchKeyEventArgs{
				Type:                  string(DispatchEventTypeRawKeyDown),
				Key:                   key,
				Code:                  code,
				NativeVirtualKeyCode:  nativeVirtualCode,
				WindowsVirtualKeyCode: nativeVirtualCode,
				KeyIdentifier:         keyIdentifier,
				IsSystemKey:           sysKey,
			})
			if err != nil {
				return nil, err
			}
		}

		err := p.client.Input.DispatchKeyEvent(ctx, &input.DispatchKeyEventArgs{
			Type:                  string(DispatchEventTypeKeyDown),
			Text:                  toType,
			UnmodifiedText:        toType,
			Key:                   key,
			Code:                  code,
			NativeVirtualKeyCode:  nativeVirtualCode,
			WindowsVirtualKeyCode: nativeVirtualCode,
			KeyIdentifier:         keyIdentifier,
			IsSystemKey:           sysKey,
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
			IsSystemKey:           sysKey,
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

// PagePressKeyInput specifies input for the PressKey method.
type PagePressKeyInput struct {
	Key        Key           // The key to press.
	Modifiers  int           // Bit field for modifier keys. Combine with |, e.g. ModifierCtrl | ModifierShift.
	Count      int           // Number of times to press the key (default: 1 if 0).
	AutoRepeat bool          // If true, simulates holding the key (event.repeat=true). If false (default), simulates discrete presses.
	Delay      time.Duration // (optional) Duration between repeated key presses.
	DelayFunc  TypeDelayFunc // (optional) Custom function for delays between repeated key presses.
}

// PagePressKeyOutput represents the output of the PressKey method.
type PagePressKeyOutput struct{}

// PressKey dispatches key press events for special keys and key combinations.
func (p *page) PressKey(ctx context.Context, in *PagePressKeyInput) (*PagePressKeyOutput, error) {
	count := in.Count
	if count <= 0 {
		count = 1
	}

	key := in.Key.Key
	code := in.Key.Code
	keyCode := in.Key.KeyCode

	for i := 0; i < count; i++ {
		isRepeat := in.AutoRepeat && i > 0

		// 1. Dispatch rawKeyDown
		args := &input.DispatchKeyEventArgs{
			Type:                  string(DispatchEventTypeRawKeyDown),
			Key:                   &key,
			Code:                  &code,
			WindowsVirtualKeyCode: &keyCode,
			NativeVirtualKeyCode:  &keyCode,
		}
		if in.Modifiers != 0 {
			args.Modifiers = &in.Modifiers
		}
		if isRepeat {
			args.AutoRepeat = &isRepeat
		}
		if err := p.client.Input.DispatchKeyEvent(ctx, args); err != nil {
			return nil, err
		}

		// 2. Dispatch char event if the key produces text
		if in.Key.Text != "" {
			text := in.Key.Text
			if in.Modifiers&ModifierShift != 0 {
				text = strings.ToUpper(text)
			}
			charArgs := &input.DispatchKeyEventArgs{
				Type:                  string(DispatchEventTypeChar),
				Key:                   &key,
				Code:                  &code,
				Text:                  &text,
				UnmodifiedText:        &text,
				WindowsVirtualKeyCode: &keyCode,
				NativeVirtualKeyCode:  &keyCode,
			}
			if in.Modifiers != 0 {
				charArgs.Modifiers = &in.Modifiers
			}
			if err := p.client.Input.DispatchKeyEvent(ctx, charArgs); err != nil {
				return nil, err
			}
		}

		// 3. Dispatch keyUp
		// AutoRepeat: only send keyUp on the very last iteration (simulates holding the key)
		// Manual repeat: send keyUp every iteration (simulates discrete presses)
		if !in.AutoRepeat || i == count-1 {
			upArgs := &input.DispatchKeyEventArgs{
				Type:                  string(DispatchEventTypeKeyUp),
				Key:                   &key,
				Code:                  &code,
				WindowsVirtualKeyCode: &keyCode,
				NativeVirtualKeyCode:  &keyCode,
			}
			if in.Modifiers != 0 {
				upArgs.Modifiers = &in.Modifiers
			}
			if err := p.client.Input.DispatchKeyEvent(ctx, upArgs); err != nil {
				return nil, err
			}
		}

		// Delay between repeated presses (skip after last)
		if i < count-1 {
			if in.Delay > 0 {
				if err := sleepWithCtx(ctx, in.Delay); err != nil {
					return nil, err
				}
			} else if in.DelayFunc != nil {
				if err := sleepWithCtx(ctx, in.DelayFunc()); err != nil {
					return nil, err
				}
			}
		}
	}

	return &PagePressKeyOutput{}, nil
}
