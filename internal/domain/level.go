package domain

// Level is one cumulative difficulty step of the trainer.
// Levels are cumulative: the Allowed set of level N contains every character
// introduced up to and including N.
type Level struct {
	Number   int
	Name     string
	NewChars []rune // characters introduced at this level (teaching order)
	allowed  map[rune]bool
}

// Allows reports whether the character is part of this level's practice set.
func (l Level) Allows(ch rune) bool { return l.allowed[ch] }

// AllowedString returns the cumulative allowed characters as a string.
func (l Level) AllowedString() string {
	// Deterministic order: sort by rune value.
	runes := make([]rune, 0, len(l.allowed))
	for ch := range l.allowed {
		runes = append(runes, ch)
	}
	sortRunes(runes)
	return string(runes)
}

type levelGroup struct {
	number  int
	name    string
	row     string
	from    int // inclusive key index
	to      int // inclusive key index
	shift   ShiftState
	accents []rune // extra dead-key characters introduced at this level
}

// levelGroups defines, per level, which physical keys (by row + index range +
// shift state) are introduced. Index ranges follow the layout.yaml key order
// left to right. The middle/upper rows use indices 0..9 (their two rightmost
// keys are punctuation/accent keys handled separately); bottom and numeric use
// the full row.
var levelGroups = []levelGroup{
	{1, "Fila media minúsculas", "middle", 0, 9, Normal, nil},
	{2, "Fila media mayúsculas", "middle", 0, 9, Shift, nil},
	{3, "Fila alta minúsculas", "upper", 0, 9, Normal, []rune("áéíóú")},
	{4, "Fila alta mayúsculas", "upper", 0, 9, Shift, []rune("ÁÉÍÓÚ")},
	{5, "Fila baja minúsculas", "bottom", 0, 10, Normal, nil},
	{6, "Fila baja mayúsculas", "bottom", 0, 10, Shift, nil},
	{7, "Fila numerada mayúsculas", "numeric", 0, 12, Shift, nil},
	{8, "Fila numerada minúsculas", "numeric", 0, 12, Normal, nil},
}

// BuildLevels derives the cumulative levels from the keyboard layout.
func BuildLevels(kb Keyboard) []Level {
	allowed := map[rune]bool{}
	levels := make([]Level, 0, len(levelGroups))

	for _, g := range levelGroups {
		var newChars []rune
		if row := kb.Row(g.row); row != nil {
			for i := g.from; i <= g.to && i < len(row.Keys); i++ {
				if ch := row.Keys[i].Char(g.shift); ch != 0 {
					newChars = append(newChars, ch)
				}
			}
		}
		newChars = append(newChars, g.accents...)

		for _, ch := range newChars {
			allowed[ch] = true
		}

		levels = append(levels, Level{
			Number:   g.number,
			Name:     g.name,
			NewChars: newChars,
			allowed:  cloneRuneSet(allowed),
		})
	}

	return levels
}

func cloneRuneSet(src map[rune]bool) map[rune]bool {
	dst := make(map[rune]bool, len(src))
	for ch, ok := range src {
		dst[ch] = ok
	}
	return dst
}

func sortRunes(r []rune) {
	for i := 1; i < len(r); i++ {
		for j := i; j > 0 && r[j] < r[j-1]; j-- {
			r[j], r[j-1] = r[j-1], r[j]
		}
	}
}
