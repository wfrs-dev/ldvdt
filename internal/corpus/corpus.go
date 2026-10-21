// Package corpus loads the embedded practice word lists and selects words for a
// given level and language.
package corpus

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	"github.com/wfrs-dev/ldvdt/internal/domain"
	"github.com/wfrs-dev/ldvdt/resources"
)

// Language is an alias of domain.Language for ergonomic call sites.
type Language = domain.Language

const (
	Spanish = domain.LanguageSpanish
	English = domain.LanguageEnglish
)

var languageFiles = map[Language]string{
	Spanish: "spa.dat",
	English: "eng.dat",
}

// Corpus holds the filtered word lists per language.
type Corpus struct {
	words map[Language][]string
}

// Load reads the embedded .dat files and filters out numeric/blank lines (S16),
// keeping only alphabetic words.
func Load() (*Corpus, error) {
	c := &Corpus{words: make(map[Language][]string, len(languageFiles))}
	for lang, file := range languageFiles {
		data, err := resources.FS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("corpus: read %s: %w", file, err)
		}
		c.words[lang] = filterWords(string(data))
	}
	return c, nil
}

// Words returns the filtered word list for the language.
func (c *Corpus) Words(lang Language) []string { return c.words[lang] }

// Select returns count practice words for the level and language. It prefers
// real corpus words whose characters are all allowed at the level and fills any
// shortfall with random 5–10 character tokens built from the level's allowed
// characters (S10, S15).
func (c *Corpus) Select(level domain.Level, lang Language, count int, rng *rand.Rand) []string {
	if count < 1 {
		return nil
	}
	eligible := c.eligible(level, lang)
	shuffle(eligible, rng)

	out := make([]string, 0, count)
	if len(eligible) >= count {
		out = append(out, eligible[:count]...)
		return out
	}

	out = append(out, eligible...)
	allowed := []rune(level.AllowedString())
	for len(out) < count {
		out = append(out, randomToken(allowed, rng))
	}
	return out
}

func (c *Corpus) eligible(level domain.Level, lang Language) []string {
	var out []string
	for _, w := range c.words[lang] {
		if allowedWord(level, w) {
			out = append(out, w)
		}
	}
	return out
}

func filterWords(data string) []string {
	var out []string
	for _, line := range strings.Split(data, "\n") {
		w := strings.TrimSpace(line)
		if w == "" || !isWord(w) {
			continue
		}
		out = append(out, w)
	}
	return out
}

// isWord reports whether s contains only letters (accented and ñ included).
func isWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func allowedWord(level domain.Level, w string) bool {
	for _, r := range w {
		if !level.Allows(r) {
			return false
		}
	}
	return true
}

func randomToken(allowed []rune, rng *rand.Rand) string {
	if len(allowed) == 0 {
		return ""
	}
	n := 5 + rng.Intn(6) // 5..10
	b := make([]rune, n)
	for i := range b {
		b[i] = allowed[rng.Intn(len(allowed))]
	}
	return string(b)
}

func shuffle(s []string, rng *rand.Rand) {
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}
