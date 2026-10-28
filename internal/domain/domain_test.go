package domain

import "testing"

func TestDefaultKeyboardShape(t *testing.T) {
	kb := DefaultKeyboard()

	wantRows := []struct {
		name string
		n    int
	}{
		{"numeric", 13},
		{"upper", 12},
		{"middle", 12},
		{"bottom", 11},
	}

	if len(kb.Rows) != len(wantRows) {
		t.Fatalf("got %d rows, want %d", len(kb.Rows), len(wantRows))
	}

	for i, w := range wantRows {
		if kb.Rows[i].Name != w.name {
			t.Errorf("row %d: got name %q, want %q", i, kb.Rows[i].Name, w.name)
		}
		if len(kb.Rows[i].Keys) != w.n {
			t.Errorf("row %q: got %d keys, want %d", w.name, len(kb.Rows[i].Keys), w.n)
		}
	}

	if got := len(kb.Keys()); got != 48 {
		t.Errorf("total keys = %d, want 48", got)
	}
}

// TestLayoutKeysHaveUniqueIDs pins the id field the SVG design relies on (S23):
// every key has a positive id and no two keys share one. The client keymap is
// exactly this mapping, so a duplicate would repaint the wrong key.
func TestLayoutKeysHaveUniqueIDs(t *testing.T) {
	seen := make(map[int]Key, 48)
	for _, k := range DefaultKeyboard().Keys() {
		if k.ID <= 0 {
			t.Errorf("key %s/%d has id %d, want a positive id", k.Row, k.Index, k.ID)
			continue
		}
		if prev, dup := seen[k.ID]; dup {
			t.Errorf("id %d is shared by %s/%d and %s/%d",
				k.ID, prev.Row, prev.Index, k.Row, k.Index)
		}
		seen[k.ID] = k
	}
	if len(seen) != 48 {
		t.Errorf("distinct ids = %d, want 48", len(seen))
	}
}

func TestFindChar(t *testing.T) {
	kb := DefaultKeyboard()

	cases := []struct {
		ch    rune
		row   string
		index int
		shift ShiftState
		ok    bool
	}{
		{'a', "middle", 0, Normal, true},
		{'A', "middle", 0, Shift, true},
		{'s', "middle", 9, Normal, true},
		{'S', "middle", 9, Shift, true},
		{'ñ', "upper", 0, Normal, true},
		{'Ñ', "upper", 0, Shift, true},
		{'p', "upper", 3, Normal, true},
		{'1', "numeric", 1, Shift, true}, // digits live on the shift plane
		{'&', "numeric", 1, Normal, true},
		{'z', "bottom", 10, Normal, true},
		{'Z', "bottom", 10, Shift, true},
		{'q', "bottom", 2, Normal, true},
		{'@', "unknown", 0, Normal, false},
	}

	for _, c := range cases {
		got, ok := kb.FindChar(c.ch)
		if ok != c.ok {
			t.Errorf("FindChar(%q) ok = %v, want %v", c.ch, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if got.Key.Row != c.row || got.Key.Index != c.index || got.Shift != c.shift {
			t.Errorf("FindChar(%q) = {row:%s idx:%d shift:%d}, want {row:%s idx:%d shift:%d}",
				c.ch, got.Key.Row, got.Key.Index, got.Shift, c.row, c.index, c.shift)
		}
	}
}

func TestBuildLevels(t *testing.T) {
	kb := DefaultKeyboard()
	levels := BuildLevels(kb)

	if len(levels) != 8 {
		t.Fatalf("got %d levels, want 8", len(levels))
	}

	type membership struct {
		level int
		ch    rune
		want  bool
	}

	cases := []membership{
		// Level 1: home row lowercase only.
		{1, 'a', true}, {1, 's', true}, {1, 'A', false}, {1, 'ñ', false}, {1, 'z', false},
		// Level 2: adds home row uppercase.
		{2, 'A', true}, {2, 'a', true}, {2, 'ñ', false},
		// Level 3: adds upper row lowercase + accents.
		{3, 'ñ', true}, {3, 'p', true}, {3, ',', true}, {3, 'á', true}, {3, 'Ñ', false}, {3, 'j', false},
		// Level 4: adds upper row shift (shift of "," is ";", of "." is ":") + uppercase accents.
		{4, 'Ñ', true}, {4, ';', true}, {4, ':', true}, {4, 'P', true}, {4, 'Á', true},
		// Level 5: adds bottom row lowercase (includes 'j').
		{5, 'j', true}, {5, 'q', true}, {5, 'z', true}, {5, '*', true}, {5, 'J', false},
		// Level 6: adds bottom row uppercase.
		{6, 'J', true}, {6, 'Q', true}, {6, 'Z', true}, {6, '%', true},
		// Level 7: adds numeric row shift (digits).
		{7, '1', true}, {7, '0', true}, {7, '?', true}, {7, '~', true}, {7, '&', false},
		// Level 8: adds numeric row normal (symbols).
		{8, '&', true}, {8, '/', true}, {8, '+', true}, {8, '$', true}, {8, '`', true},
	}

	for _, c := range cases {
		l := levels[c.level-1]
		if got := l.Allows(c.ch); got != c.want {
			t.Errorf("level %d Allows(%q) = %v, want %v", c.level, c.ch, got, c.want)
		}
	}
}

func TestLevelNewCharsOrdering(t *testing.T) {
	kb := DefaultKeyboard()
	levels := BuildLevels(kb)

	if got := string(levels[0].NewChars); got != "aoeiudrtns" {
		t.Errorf("level 1 NewChars = %q, want %q", got, "aoeiudrtns")
	}
	if got := string(levels[2].NewChars); got != "ñ,.pyfgchláéíóú" {
		t.Errorf("level 3 NewChars = %q, want %q", got, "ñ,.pyfgchláéíóú")
	}
}

func TestFindCharAccents(t *testing.T) {
	kb := DefaultKeyboard()

	cases := []struct {
		ch       rune
		base     rune
		baseSh   ShiftState
		wantDead bool
	}{
		{'á', 'a', Normal, true},
		{'é', 'e', Normal, true},
		{'í', 'i', Normal, true},
		{'ó', 'o', Normal, true},
		{'ú', 'u', Normal, true},
		{'Á', 'A', Shift, true},
		{'É', 'E', Shift, true},
		{'Í', 'I', Shift, true},
		{'Ó', 'O', Shift, true},
		{'Ú', 'U', Shift, true},
		// Direct keys must not become dead-key combinations.
		{'ñ', 'ñ', Normal, false},
		{'Ñ', 'Ñ', Shift, false},
		{'a', 'a', Normal, false},
	}

	for _, c := range cases {
		got, ok := kb.FindChar(c.ch)
		if !ok {
			t.Errorf("FindChar(%q) not found", c.ch)
			continue
		}
		if got.IsDead() != c.wantDead {
			t.Fatalf("FindChar(%q) dead = %v, want %v", c.ch, got.IsDead(), c.wantDead)
		}
		if got.Key.Char(got.Shift) != c.base || got.Shift != c.baseSh {
			t.Errorf("FindChar(%q) base = %q/%v, want %q/%v",
				c.ch, got.Key.Char(got.Shift), got.Shift, c.base, c.baseSh)
		}
		if c.wantDead {
			if d := got.DeadKey.Char(got.DeadShift); d != DeadAccent {
				t.Errorf("FindChar(%q) dead key = %q, want %q", c.ch, d, DeadAccent)
			}
		}
	}
}
