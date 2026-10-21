package ui

import (
	"bytes"
	"context"
	"io"
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
}

func TestPracticeRendersChars(t *testing.T) {
	html := render(t, Practice([]string{"isla", "otro"}, 1, "es", 2, "aoeiudrtns"))

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
}
