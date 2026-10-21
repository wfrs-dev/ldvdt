package domain

// ShiftState describes whether a character is produced with the SHIFT
// modifier held down (Top) or with a plain key press (Bottom).
type ShiftState int

const (
	// Normal is a plain key press (no modifier); it produces the key's Bottom char.
	Normal ShiftState = iota
	// Shift means SHIFT is held; it produces the key's Top char.
	Shift
)

// Key is one physical key of the ldvd layout.
type Key struct {
	Row    string // "numeric", "upper", "middle" or "bottom"
	Index  int    // 0-based position, left to right, within the row
	Top    rune   // character produced with SHIFT
	Bottom rune   // character produced without SHIFT
}

// Finger returns the finger assigned to this key.
func (k Key) Finger() Finger { return FingerFor(k.Row, k.Index) }

// Char returns the character this key produces for the given shift state.
func (k Key) Char(s ShiftState) rune {
	if s == Shift {
		return k.Top
	}
	return k.Bottom
}

// Row is a horizontal line of keys on the keyboard.
type Row struct {
	Name string
	Keys []Key
}

// Keyboard is the full fixed ldvd layout. It is generated at build time from
// docs/layout.yaml (see internal/tools/genlayout) and embedded in the binary;
// the yaml file is never read at runtime.
type Keyboard struct {
	Rows []Row
}

// Row returns the row with the given name, or nil.
func (kb Keyboard) Row(name string) *Row {
	for i := range kb.Rows {
		if kb.Rows[i].Name == name {
			return &kb.Rows[i]
		}
	}
	return nil
}

// Keys returns every key of the keyboard in display order.
func (kb Keyboard) Keys() []Key {
	var keys []Key
	for i := range kb.Rows {
		keys = append(keys, kb.Rows[i].Keys...)
	}
	return keys
}

// CharKey is the key press needed to produce a single character.
type CharKey struct {
	Key      Key
	Shift    ShiftState
	DeadKey  *Key  // non-nil when the char needs a dead key first (accents)
	DeadShift ShiftState
}

// IsDead reports whether the character requires a dead key combination.
func (c CharKey) IsDead() bool { return c.DeadKey != nil }

// FindChar returns the key press that produces the given rune, or false if the
// character is not part of the layout. It searches Bottom first (plain press)
// and then Top (shift press), so a character that appears both ways resolves to
// the plain press.
func (kb Keyboard) FindChar(ch rune) (CharKey, bool) {
	for i := range kb.Rows {
		for j := range kb.Rows[i].Keys {
			k := kb.Rows[i].Keys[j]
			if k.Bottom == ch {
				return CharKey{Key: k, Shift: Normal}, true
			}
		}
	}
	for i := range kb.Rows {
		for j := range kb.Rows[i].Keys {
			k := kb.Rows[i].Keys[j]
			if k.Top == ch {
				return CharKey{Key: k, Shift: Shift}, true
			}
		}
	}
	return CharKey{}, false
}

// DefaultKeyboard returns the fixed, generated ldvd layout.
func DefaultKeyboard() Keyboard {
	kb := layoutKeyboard
	// Finger assignment is deterministic and computed on demand; nothing to
	// mutate here. Return a copy to keep callers from mutating the embedded data.
	return kb
}
