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
	"time"

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

// renderPage renders the whole SPA shell with the default keyboard and levels,
// the fixture every page-level test needs.
func renderPage(t *testing.T) string {
	t.Helper()
	kb := domain.DefaultKeyboard()
	return render(t, Page(kb, domain.BuildLevels(kb), domain.Stats{}, "[]"))
}

func TestPageRenders(t *testing.T) {
	html := renderPage(t)

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

// TestPageStacksThreeRows pins the shell structure of S22: everything lives
// inside a single Bootstrap .container, and the three rows keep their order —
// controls, then the words and keyboard, then the progress panel. Document order
// is what the user sees, because the rows stack.
func TestPageStacksThreeRows(t *testing.T) {
	html := renderPage(t)

	container := strings.Index(html, `<div class="container`)
	if container < 0 {
		t.Fatal("page is not wrapped in a Bootstrap container")
	}

	// Each row is a Bootstrap .row carrying its own id, so the grid classes and
	// the document order can be checked together.
	rows := []string{"controls-row", "practice-row", "progress-panel"}
	previous := container
	for _, id := range rows {
		tag, at := tagOf(t, html, `id="`+id+`"`)
		if !strings.Contains(tag, `class="row`) {
			t.Errorf("row %q is not a Bootstrap row: %s", id, tag)
		}
		if at < previous {
			t.Errorf("row %q at %d must come after the previous region at %d", id, at, previous)
		}
		previous = at
	}

	// The configuration form belongs to the first row.
	form := strings.Index(html, `id="controls"`)
	words := strings.Index(html, `id="words"`)
	if form < container || words < container {
		t.Fatalf("form or words outside the container: form=%d words=%d container=%d",
			form, words, container)
	}
	if form > words {
		t.Errorf("controls form at %d must precede the words at %d", form, words)
	}

	// The retired layout classes must not linger.
	for _, gone := range []string{"panel-left", "panel-right", "class=\"app\"", "panel-controls", "panel-progress"} {
		if strings.Contains(html, gone) {
			t.Errorf("page still uses the retired layout class %q", gone)
		}
	}
}

// TestPageUsesBootstrapComponents pins that the interactive parts are real
// Bootstrap controls (S21), not app CSS pretending to be one: without the
// component classes the themed build never reaches the form.
func TestPageUsesBootstrapComponents(t *testing.T) {
	html := renderPage(t)

	for _, want := range []string{
		`class="form-select`,
		`class="form-control`,
		`class="btn btn-primary`,
		`class="btn btn-outline-secondary`,
		`class="form-label`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page HTML missing Bootstrap component %q", want)
		}
	}
}

// tagOf returns the opening tag that carries marker and its offset, so a test
// can assert the attributes of the element with a given id without an HTML
// parser.
func tagOf(t *testing.T, html, marker string) (string, int) {
	t.Helper()
	at := strings.Index(html, marker)
	if at < 0 {
		t.Fatalf("marker %q not found", marker)
	}
	start := strings.LastIndex(html[:at], "<")
	end := strings.Index(html[at:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("marker %q is not inside a tag", marker)
	}
	return html[start : at+end+1], at
}

// TestPageRendersThemeToggleAndBootstrap pins the theme plumbing the client
// needs (S20, S24): the manual toggle in the first row, the Bootstrap theme
// attribute, the inline head script that applies the persisted theme before
// first paint, and the vendored Bootstrap stylesheet loading before app.css.
func TestPageRendersThemeToggleAndBootstrap(t *testing.T) {
	html := renderPage(t)

	if !strings.Contains(html, `id="theme-toggle"`) {
		t.Error("page missing the theme toggle button")
	}
	if !strings.Contains(html, `data-bs-theme="dark"`) {
		t.Error("page missing the dark default data-bs-theme attribute")
	}
	if !strings.Contains(html, `localStorage.getItem("ldvdt-theme")`) {
		t.Error("inline bootstrap must read the ldvdt-theme localStorage key")
	}
	if !strings.Contains(html, `setAttribute("data-bs-theme", theme)`) {
		t.Error("inline bootstrap must apply the saved theme before first paint")
	}

	// The bootstrap has to run before the stylesheets, otherwise the first paint
	// can use the wrong palette. That ordering is the whole point of the script,
	// so it is pinned instead of the exact source text.
	bootstrap := strings.Index(html, `localStorage.getItem("ldvdt-theme")`)
	bootstrapCSS := strings.Index(html, `href="/static/vendor/bootstrap.min.css"`)
	appCSS := strings.Index(html, `href="/static/app.css"`)
	if bootstrap < 0 || bootstrapCSS < 0 || appCSS < 0 {
		t.Fatalf("missing link: bootstrap=%d bootstrapCSS=%d appCSS=%d", bootstrap, bootstrapCSS, appCSS)
	}
	if bootstrap > bootstrapCSS {
		t.Errorf("theme bootstrap at %d must precede the stylesheets at %d", bootstrap, bootstrapCSS)
	}
	// Bootstrap owns the palette and the components (S21); app.css only adds
	// what Bootstrap does not have, so it has to load after it.
	if bootstrapCSS > appCSS {
		t.Errorf("bootstrap CSS at %d must precede app.css at %d", bootstrapCSS, appCSS)
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

func TestSeriesJSON(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

	got, err := SeriesJSON(nil, now)
	if err != nil {
		t.Fatalf("SeriesJSON(nil) error: %v", err)
	}
	if got != "[]" {
		t.Errorf("SeriesJSON(nil) = %q, want []", got)
	}

	practices := []domain.Practice{
		// Yesterday: the label carries the full date and 12-hour time so points
		// from different days (and years) stay distinguishable on the axis.
		{CompletedAt: time.Date(2026, 10, 19, 23, 30, 0, 0, time.UTC), WPM: 22.46, Accuracy: 0.9149, Consistency: 30},
		// Today: 12-hour time only.
		{CompletedAt: time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC), WPM: 28.0, Accuracy: 0.955, Consistency: 25},
	}

	got, err = SeriesJSON(practices, now)
	if err != nil {
		t.Fatalf("SeriesJSON error: %v", err)
	}
	want := `[{"label":"19/10/2026 11:30PM","wpm":22.5,"acc":91.5},{"label":"10:00AM","wpm":28,"acc":95.5}]`
	if got != want {
		t.Errorf("SeriesJSON =\n %s\nwant\n %s", got, want)
	}
}

// TestSeriesJSONNormalisesToNowZone pins that a practice timestamp is rendered
// in now's location, not in the location it happened to be constructed with.
// The "today" boundary and the clock must agree even when at and now carry
// different zones (e.g. a UTC value read back next to a local now).
func TestSeriesJSONNormalisesToNowZone(t *testing.T) {
	loc := time.FixedZone("UTC-5", -5*3600)
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, loc)

	// 2026-10-19 23:30 UTC is 2026-10-19 18:30 in UTC-5: the previous local day.
	at := time.Date(2026, 10, 19, 23, 30, 0, 0, time.UTC)

	got, err := SeriesJSON([]domain.Practice{
		{CompletedAt: at, WPM: 22.46, Accuracy: 0.9149, Consistency: 30},
	}, now)
	if err != nil {
		t.Fatalf("SeriesJSON error: %v", err)
	}
	want := `[{"label":"19/10/2026 06:30PM","wpm":22.5,"acc":91.5}]`
	if got != want {
		t.Errorf("SeriesJSON =\n %s\nwant\n %s", got, want)
	}
}

func TestProgressRendersSeries(t *testing.T) {
	html := render(t, Progress(domain.Stats{}, `[{"label":"10:00","wpm":20,"acc":90}]`))

	if !strings.Contains(html, `id="progress-chart"`) {
		t.Error("missing progress-chart canvas")
	}
	if !strings.Contains(html, `data-series=`) {
		t.Error("missing data-series attribute")
	}
}
