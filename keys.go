package gopilot

// Key represents a keyboard key with its CDP protocol properties.
type Key struct {
	Key     string // DOM key value (e.g., "Enter", "Tab", "a")
	Code    string // Physical key code (e.g., "Enter", "Tab", "KeyA")
	KeyCode int    // Windows virtual key code (e.g., 13 for Enter)
	Text    string // Text produced by the key (e.g., "\r" for Enter, "" for Tab)
}

// Modifier constants for use with PagePressKeyInput.Modifiers.
// Combine with bitwise OR: ModifierCtrl | ModifierShift.
// Maps directly to CDP DispatchKeyEventArgs.Modifiers bit field.
const (
	ModifierAlt   = 1
	ModifierCtrl  = 2
	ModifierMeta  = 4
	ModifierShift = 8
)

// Special keys
var (
	KeyEnter     = Key{Key: "Enter", Code: "Enter", KeyCode: 13, Text: "\r"}
	KeyTab       = Key{Key: "Tab", Code: "Tab", KeyCode: 9}
	KeyEscape    = Key{Key: "Escape", Code: "Escape", KeyCode: 27}
	KeyBackspace = Key{Key: "Backspace", Code: "Backspace", KeyCode: 8}
	KeyDelete    = Key{Key: "Delete", Code: "Delete", KeyCode: 46}
	KeyInsert    = Key{Key: "Insert", Code: "Insert", KeyCode: 45}
	KeySpace     = Key{Key: " ", Code: "Space", KeyCode: 32, Text: " "}
)

// Arrow keys
var (
	KeyArrowUp    = Key{Key: "ArrowUp", Code: "ArrowUp", KeyCode: 38}
	KeyArrowDown  = Key{Key: "ArrowDown", Code: "ArrowDown", KeyCode: 40}
	KeyArrowLeft  = Key{Key: "ArrowLeft", Code: "ArrowLeft", KeyCode: 37}
	KeyArrowRight = Key{Key: "ArrowRight", Code: "ArrowRight", KeyCode: 39}
)

// Page navigation keys
var (
	KeyHome     = Key{Key: "Home", Code: "Home", KeyCode: 36}
	KeyEnd      = Key{Key: "End", Code: "End", KeyCode: 35}
	KeyPageUp   = Key{Key: "PageUp", Code: "PageUp", KeyCode: 33}
	KeyPageDown = Key{Key: "PageDown", Code: "PageDown", KeyCode: 34}
)

// Function keys
var (
	KeyF1  = Key{Key: "F1", Code: "F1", KeyCode: 112}
	KeyF2  = Key{Key: "F2", Code: "F2", KeyCode: 113}
	KeyF3  = Key{Key: "F3", Code: "F3", KeyCode: 114}
	KeyF4  = Key{Key: "F4", Code: "F4", KeyCode: 115}
	KeyF5  = Key{Key: "F5", Code: "F5", KeyCode: 116}
	KeyF6  = Key{Key: "F6", Code: "F6", KeyCode: 117}
	KeyF7  = Key{Key: "F7", Code: "F7", KeyCode: 118}
	KeyF8  = Key{Key: "F8", Code: "F8", KeyCode: 119}
	KeyF9  = Key{Key: "F9", Code: "F9", KeyCode: 120}
	KeyF10 = Key{Key: "F10", Code: "F10", KeyCode: 121}
	KeyF11 = Key{Key: "F11", Code: "F11", KeyCode: 122}
	KeyF12 = Key{Key: "F12", Code: "F12", KeyCode: 123}
)

// Letter keys
var (
	KeyA = Key{Key: "a", Code: "KeyA", KeyCode: 65, Text: "a"}
	KeyB = Key{Key: "b", Code: "KeyB", KeyCode: 66, Text: "b"}
	KeyC = Key{Key: "c", Code: "KeyC", KeyCode: 67, Text: "c"}
	KeyD = Key{Key: "d", Code: "KeyD", KeyCode: 68, Text: "d"}
	KeyE = Key{Key: "e", Code: "KeyE", KeyCode: 69, Text: "e"}
	KeyF = Key{Key: "f", Code: "KeyF", KeyCode: 70, Text: "f"}
	KeyG = Key{Key: "g", Code: "KeyG", KeyCode: 71, Text: "g"}
	KeyH = Key{Key: "h", Code: "KeyH", KeyCode: 72, Text: "h"}
	KeyI = Key{Key: "i", Code: "KeyI", KeyCode: 73, Text: "i"}
	KeyJ = Key{Key: "j", Code: "KeyJ", KeyCode: 74, Text: "j"}
	KeyK = Key{Key: "k", Code: "KeyK", KeyCode: 75, Text: "k"}
	KeyL = Key{Key: "l", Code: "KeyL", KeyCode: 76, Text: "l"}
	KeyM = Key{Key: "m", Code: "KeyM", KeyCode: 77, Text: "m"}
	KeyN = Key{Key: "n", Code: "KeyN", KeyCode: 78, Text: "n"}
	KeyO = Key{Key: "o", Code: "KeyO", KeyCode: 79, Text: "o"}
	KeyP = Key{Key: "p", Code: "KeyP", KeyCode: 80, Text: "p"}
	KeyQ = Key{Key: "q", Code: "KeyQ", KeyCode: 81, Text: "q"}
	KeyR = Key{Key: "r", Code: "KeyR", KeyCode: 82, Text: "r"}
	KeyS = Key{Key: "s", Code: "KeyS", KeyCode: 83, Text: "s"}
	KeyT = Key{Key: "t", Code: "KeyT", KeyCode: 84, Text: "t"}
	KeyU = Key{Key: "u", Code: "KeyU", KeyCode: 85, Text: "u"}
	KeyV = Key{Key: "v", Code: "KeyV", KeyCode: 86, Text: "v"}
	KeyW = Key{Key: "w", Code: "KeyW", KeyCode: 87, Text: "w"}
	KeyX = Key{Key: "x", Code: "KeyX", KeyCode: 88, Text: "x"}
	KeyY = Key{Key: "y", Code: "KeyY", KeyCode: 89, Text: "y"}
	KeyZ = Key{Key: "z", Code: "KeyZ", KeyCode: 90, Text: "z"}
)

// Digit keys
var (
	Key0 = Key{Key: "0", Code: "Digit0", KeyCode: 48, Text: "0"}
	Key1 = Key{Key: "1", Code: "Digit1", KeyCode: 49, Text: "1"}
	Key2 = Key{Key: "2", Code: "Digit2", KeyCode: 50, Text: "2"}
	Key3 = Key{Key: "3", Code: "Digit3", KeyCode: 51, Text: "3"}
	Key4 = Key{Key: "4", Code: "Digit4", KeyCode: 52, Text: "4"}
	Key5 = Key{Key: "5", Code: "Digit5", KeyCode: 53, Text: "5"}
	Key6 = Key{Key: "6", Code: "Digit6", KeyCode: 54, Text: "6"}
	Key7 = Key{Key: "7", Code: "Digit7", KeyCode: 55, Text: "7"}
	Key8 = Key{Key: "8", Code: "Digit8", KeyCode: 56, Text: "8"}
	Key9 = Key{Key: "9", Code: "Digit9", KeyCode: 57, Text: "9"}
)

// Symbol keys (US layout)
var (
	KeyMinus        = Key{Key: "-", Code: "Minus", KeyCode: 189, Text: "-"}
	KeyEqual        = Key{Key: "=", Code: "Equal", KeyCode: 187, Text: "="}
	KeyBracketLeft  = Key{Key: "[", Code: "BracketLeft", KeyCode: 219, Text: "["}
	KeyBracketRight = Key{Key: "]", Code: "BracketRight", KeyCode: 221, Text: "]"}
	KeyBackslash    = Key{Key: "\\", Code: "Backslash", KeyCode: 220, Text: "\\"}
	KeySemicolon    = Key{Key: ";", Code: "Semicolon", KeyCode: 186, Text: ";"}
	KeyQuote        = Key{Key: "'", Code: "Quote", KeyCode: 222, Text: "'"}
	KeyBackquote    = Key{Key: "`", Code: "Backquote", KeyCode: 192, Text: "`"}
	KeyComma        = Key{Key: ",", Code: "Comma", KeyCode: 188, Text: ","}
	KeyPeriod       = Key{Key: ".", Code: "Period", KeyCode: 190, Text: "."}
	KeySlash        = Key{Key: "/", Code: "Slash", KeyCode: 191, Text: "/"}
)

// runeToKey maps runes to their PagePressKeyInput (US keyboard layout).
var runeToKey = map[rune]PagePressKeyInput{
	// Letters
	'a': {Key: KeyA}, 'b': {Key: KeyB}, 'c': {Key: KeyC}, 'd': {Key: KeyD},
	'e': {Key: KeyE}, 'f': {Key: KeyF}, 'g': {Key: KeyG}, 'h': {Key: KeyH},
	'i': {Key: KeyI}, 'j': {Key: KeyJ}, 'k': {Key: KeyK}, 'l': {Key: KeyL},
	'm': {Key: KeyM}, 'n': {Key: KeyN}, 'o': {Key: KeyO}, 'p': {Key: KeyP},
	'q': {Key: KeyQ}, 'r': {Key: KeyR}, 's': {Key: KeyS}, 't': {Key: KeyT},
	'u': {Key: KeyU}, 'v': {Key: KeyV}, 'w': {Key: KeyW}, 'x': {Key: KeyX},
	'y': {Key: KeyY}, 'z': {Key: KeyZ},

	// Uppercase letters
	'A': {Key: KeyA, Modifiers: ModifierShift}, 'B': {Key: KeyB, Modifiers: ModifierShift},
	'C': {Key: KeyC, Modifiers: ModifierShift}, 'D': {Key: KeyD, Modifiers: ModifierShift},
	'E': {Key: KeyE, Modifiers: ModifierShift}, 'F': {Key: KeyF, Modifiers: ModifierShift},
	'G': {Key: KeyG, Modifiers: ModifierShift}, 'H': {Key: KeyH, Modifiers: ModifierShift},
	'I': {Key: KeyI, Modifiers: ModifierShift}, 'J': {Key: KeyJ, Modifiers: ModifierShift},
	'K': {Key: KeyK, Modifiers: ModifierShift}, 'L': {Key: KeyL, Modifiers: ModifierShift},
	'M': {Key: KeyM, Modifiers: ModifierShift}, 'N': {Key: KeyN, Modifiers: ModifierShift},
	'O': {Key: KeyO, Modifiers: ModifierShift}, 'P': {Key: KeyP, Modifiers: ModifierShift},
	'Q': {Key: KeyQ, Modifiers: ModifierShift}, 'R': {Key: KeyR, Modifiers: ModifierShift},
	'S': {Key: KeyS, Modifiers: ModifierShift}, 'T': {Key: KeyT, Modifiers: ModifierShift},
	'U': {Key: KeyU, Modifiers: ModifierShift}, 'V': {Key: KeyV, Modifiers: ModifierShift},
	'W': {Key: KeyW, Modifiers: ModifierShift}, 'X': {Key: KeyX, Modifiers: ModifierShift},
	'Y': {Key: KeyY, Modifiers: ModifierShift}, 'Z': {Key: KeyZ, Modifiers: ModifierShift},

	// Digits
	'0': {Key: Key0}, '1': {Key: Key1}, '2': {Key: Key2}, '3': {Key: Key3},
	'4': {Key: Key4}, '5': {Key: Key5}, '6': {Key: Key6}, '7': {Key: Key7},
	'8': {Key: Key8}, '9': {Key: Key9},

	// Whitespace
	' ':  {Key: KeySpace},
	'\t': {Key: KeyTab},
	'\r': {Key: KeyEnter},
	'\n': {Key: KeyEnter},

	// Symbols (unshifted)
	'-':  {Key: KeyMinus},
	'=':  {Key: KeyEqual},
	'[':  {Key: KeyBracketLeft},
	']':  {Key: KeyBracketRight},
	'\\': {Key: KeyBackslash},
	';':  {Key: KeySemicolon},
	'\'': {Key: KeyQuote},
	'`':  {Key: KeyBackquote},
	',':  {Key: KeyComma},
	'.':  {Key: KeyPeriod},
	'/':  {Key: KeySlash},

	// Symbols (shifted)
	'!': {Key: Key1, Modifiers: ModifierShift},
	'@': {Key: Key2, Modifiers: ModifierShift},
	'#': {Key: Key3, Modifiers: ModifierShift},
	'$': {Key: Key4, Modifiers: ModifierShift},
	'%': {Key: Key5, Modifiers: ModifierShift},
	'^': {Key: Key6, Modifiers: ModifierShift},
	'&': {Key: Key7, Modifiers: ModifierShift},
	'*': {Key: Key8, Modifiers: ModifierShift},
	'(': {Key: Key9, Modifiers: ModifierShift},
	')': {Key: Key0, Modifiers: ModifierShift},
	'_': {Key: KeyMinus, Modifiers: ModifierShift},
	'+': {Key: KeyEqual, Modifiers: ModifierShift},
	'{': {Key: KeyBracketLeft, Modifiers: ModifierShift},
	'}': {Key: KeyBracketRight, Modifiers: ModifierShift},
	'|': {Key: KeyBackslash, Modifiers: ModifierShift},
	':': {Key: KeySemicolon, Modifiers: ModifierShift},
	'"': {Key: KeyQuote, Modifiers: ModifierShift},
	'~': {Key: KeyBackquote, Modifiers: ModifierShift},
	'<': {Key: KeyComma, Modifiers: ModifierShift},
	'>': {Key: KeyPeriod, Modifiers: ModifierShift},
	'?': {Key: KeySlash, Modifiers: ModifierShift},
}

// KeyFromRune returns a PagePressKeyInput for a given rune (US keyboard layout).
// For unknown runes, it returns a fallback with just the Text field set.
func KeyFromRune(r rune) PagePressKeyInput {
	if ki, ok := runeToKey[r]; ok {
		return ki
	}
	// Fallback for unmapped runes
	t := string(r)
	return PagePressKeyInput{
		Key: Key{Key: t, Text: t},
	}
}
