package ui

import (
	"bytes"
	"context"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

func render(t *testing.T, c interface {
	Render(context.Context, io.Writer) error
}) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return buf.String()
}

func TestPageRenders(t *testing.T) {
	kb := domain.DefaultKeyboard()
	levels := domain.BuildLevels(kb)

	html := render(t, Page(kb, levels, domain.Stats{}))

	for _, want := range []string{
		"<div class=\"keyboard\"",
		"data-bottom=",
		"finger-LP",
		"id=\"controls\"",
		"id=\"words\"",
		"Progreso",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page HTML missing %q", want)
		}
	}
}

func TestKeyboardHasAllKeys(t *testing.T) {
	kb := domain.DefaultKeyboard()
	html := render(t, Keyboard(kb))

	// 48 layout keys rendered.
	if got := strings.Count(html, "class=\"key finger-"); got != 48 {
		t.Errorf("rendered %d finger keys, want 48", got)
	}

	// The spacebar must expose data-bottom=" " so the typing engine highlights it.
	if !strings.Contains(html, `data-bottom=" "`) {
		t.Error("space key missing data-bottom for highlighting")
	}

	// Decorative modifier keys for a realistic ISO look.
	for _, cls := range []string{
		"key-tab", "key-caps", "key-shift-left", "key-shift-right",
		"key-backspace", "key-enter-top", "key-enter-bottom",
	} {
		if !strings.Contains(html, cls) {
			t.Errorf("missing decorative key %q", cls)
		}
	}

	// The super (Meta/Windows) key appears on both sides of the spacebar.
	if got := strings.Count(html, "key-super"); got != 2 {
		t.Errorf("super keys = %d, want 2", got)
	}
}

// TestKeyboardRowsSameWidth enforces that every row sums to the same number of
// width units, which is what keeps both edges of the keyboard aligned.
func TestKeyboardRowsSameWidth(t *testing.T) {
	kb := domain.DefaultKeyboard()
	html := render(t, Keyboard(kb))

	const wantUnits = 15.0

	chunks := strings.Split(html, `class="keyrow `)
	if len(chunks) < 2 {
		t.Fatal("no keyboard rows rendered")
	}

	widthRe := regexp.MustCompile(`data-w="([0-9.]+)"`)
	seen := 0
	for _, chunk := range chunks[1:] {
		name := chunk[:strings.Index(chunk, `"`)]
		sum := 0.0
		for _, m := range widthRe.FindAllStringSubmatch(chunk, -1) {
			v, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				t.Fatalf("row %q: bad data-w %q: %v", name, m[1], err)
			}
			sum += v
		}
		if math.Abs(sum-wantUnits) > 1e-9 {
			t.Errorf("row %q sums to %v units, want %v", name, sum, wantUnits)
		}
		seen++
	}

	if seen != 5 {
		t.Errorf("checked %d rows, want 5", seen)
	}
}

func TestPracticeRendersChars(t *testing.T) {
	html := render(t, Practice([]string{"isla", "otro"}, 1, "es", 2, "aoeiudrtns", ""))

	if !strings.Contains(html, `data-level="1"`) {
		t.Error("missing data-level")
	}
	if !strings.Contains(html, `data-lang="es"`) {
		t.Error("missing data-lang")
	}
	if !strings.Contains(html, `<span class="char">i</span>`) {
		t.Error("missing per-char span")
	}
	if !strings.Contains(html, `data-keys="aoeiudrtns"`) {
		t.Error("missing data-keys")
	}
	if !strings.Contains(html, `data-accents=`) {
		t.Error("missing data-accents attribute")
	}
	// Two words render as two distinct word spans (visual spacing/wrapping is CSS).
	if got := strings.Count(html, `<span class="word">`); got != 2 {
		t.Errorf("word spans = %d, want 2", got)
	}
}
