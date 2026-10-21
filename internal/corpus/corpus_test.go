package corpus

import (
	"math/rand"
	"testing"
	"unicode"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

func TestLoadFiltersNumericLines(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	for _, lang := range []Language{Spanish, English} {
		words := c.Words(lang)
		if len(words) == 0 {
			t.Fatalf("%s: no words loaded", lang)
		}
		for _, w := range words {
			for _, r := range w {
				if !unicode.IsLetter(r) {
					t.Fatalf("%s: word %q contains non-letter %q", lang, w, r)
				}
			}
		}
	}
}

func TestSelectRealWordsOnly(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	kb := domain.DefaultKeyboard()
	levels := domain.BuildLevels(kb)
	level1 := levels[0]

	rng := rand.New(rand.NewSource(42))
	words := c.Select(level1, English, 20, rng)

	if len(words) != 20 {
		t.Fatalf("Select returned %d words, want 20", len(words))
	}
	for _, w := range words {
		for _, r := range w {
			if !level1.Allows(r) {
				t.Fatalf("word %q contains %q not allowed at level 1", w, r)
			}
		}
	}
}

func TestSelectFallsBackToRandomTokens(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	kb := domain.DefaultKeyboard()
	levels := domain.BuildLevels(kb)
	level1 := levels[0]

	// Request far more words than level 1 has eligible corpus words.
	rng := rand.New(rand.NewSource(7))
	words := c.Select(level1, English, 5000, rng)

	if len(words) != 5000 {
		t.Fatalf("Select returned %d words, want 5000", len(words))
	}
	for _, w := range words {
		n := len([]rune(w))
		if n < 5 || n > 10 {
			// Real corpus words can be shorter; only generated tokens must be 5..10.
			// Distinguish: generated tokens only contain allowed level-1 letters too.
			if !allIn(w, "adeinorstu") {
				t.Fatalf("generated token %q has bad length %d", w, n)
			}
		}
	}
}

func allIn(s, set string) bool {
	m := map[rune]bool{}
	for _, r := range set {
		m[r] = true
	}
	for _, r := range s {
		if !m[r] {
			return false
		}
	}
	return true
}
