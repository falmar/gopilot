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
	for i, r := range in.Text {
		ki := KeyFromRune(r)
		if err := p.dispatchKeyPress(ctx, ki.Key, ki.Modifiers, false, false); err != nil {
			return nil, err
		}

		if i == last-1 {
			break
		}

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

	for i := 0; i < count; i++ {
		isRepeat := in.AutoRepeat && i > 0

		// AutoRepeat: skip keyUp on intermediate iterations (simulates holding the key)
		skipKeyUp := in.AutoRepeat && i < count-1

		if err := p.dispatchKeyPress(ctx, in.Key, in.Modifiers, isRepeat, skipKeyUp); err != nil {
			return nil, err
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

// dispatchKeyPress dispatches the rawKeyDown → char → keyUp sequence for a single key press.
func (p *page) dispatchKeyPress(ctx context.Context, k Key, modifiers int, autoRepeat bool, skipKeyUp bool) error {
	key := k.Key
	code := k.Code
	keyCode := k.KeyCode

	// 1. Dispatch rawKeyDown
	args := &input.DispatchKeyEventArgs{
		Type:                  string(DispatchEventTypeRawKeyDown),
		Key:                   &key,
		Code:                  &code,
		WindowsVirtualKeyCode: &keyCode,
		NativeVirtualKeyCode:  &keyCode,
	}
	if modifiers != 0 {
		args.Modifiers = &modifiers
	}
	if autoRepeat {
		args.AutoRepeat = &autoRepeat
	}
	if err := p.client.Input.DispatchKeyEvent(ctx, args); err != nil {
		return err
	}

	// 2. Dispatch char event if the key produces text
	if k.Text != "" {
		text := k.Text
		if modifiers&ModifierShift != 0 {
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
		if modifiers != 0 {
			charArgs.Modifiers = &modifiers
		}
		if err := p.client.Input.DispatchKeyEvent(ctx, charArgs); err != nil {
			return err
		}
	}

	// 3. Dispatch keyUp (skipped when simulating held key on intermediate presses)
	if skipKeyUp {
		return nil
	}
	upArgs := &input.DispatchKeyEventArgs{
		Type:                  string(DispatchEventTypeKeyUp),
		Key:                   &key,
		Code:                  &code,
		WindowsVirtualKeyCode: &keyCode,
		NativeVirtualKeyCode:  &keyCode,
	}
	if modifiers != 0 {
		upArgs.Modifiers = &modifiers
	}
	if err := p.client.Input.DispatchKeyEvent(ctx, upArgs); err != nil {
		return err
	}

	return nil
}
