package trainer_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/wfrs-dev/ldvdt/internal/corpus"
	"github.com/wfrs-dev/ldvdt/internal/domain"
	"github.com/wfrs-dev/ldvdt/internal/infra/storage/sqlite"
	"github.com/wfrs-dev/ldvdt/internal/trainer"
)

func newTestServer(t *testing.T) *echo.Echo {
	t.Helper()

	store, err := sqlite.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	c, err := corpus.Load()
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}

	kb := domain.DefaultKeyboard()
	levels := domain.BuildLevels(kb)
	svc := trainer.NewService(store, c, levels)

	e := echo.New()
	trainer.NewModule(svc, kb).Register(e)
	return e
}

func do(e *echo.Echo, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestIndexRendersSPA(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	html := rec.Body.String()
	for _, want := range []string{"keyboard", "controls", "words", "Progreso"} {
		if !strings.Contains(html, want) {
			t.Errorf("GET / missing %q", want)
		}
	}

	// The served shell has to carry the themed Bootstrap sheet and the theme
	// attribute the client toggles (S21, S22, S24); the unit tests pin the
	// ordering, this one pins that the route actually serves it.
	for _, want := range []string{
		`href="/static/vendor/bootstrap.min.css"`,
		`data-bs-theme="dark"`,
		`class="container`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("GET / missing %q", want)
		}
	}
}

func TestPracticeLevel1(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/practice?level=1&lang=es&count=20", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /practice status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	html := rec.Body.String()
	if !strings.Contains(html, `data-level="1"`) {
		t.Error("missing data-level=1")
	}
	if !strings.Contains(html, `data-lang="es"`) {
		t.Error("missing data-lang=es")
	}
	if !strings.Contains(html, `<span class="word">`) {
		t.Error("missing word spans")
	}
}

func TestSubmitAndProgress(t *testing.T) {
	e := newTestServer(t)

	body := `{"level":1,"lang":"es","wordCount":20,"wpm":30.5,"accuracy":0.95,"consistency":40.0,"durationMS":60000}`
	rec := do(e, http.MethodPost, "/practice", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /practice status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Saved bool `json:"saved"`
		Stats struct {
			TotalPractices int `json:"TotalPractices"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Saved {
		t.Error("expected saved=true")
	}
	if resp.Stats.TotalPractices != 1 {
		t.Errorf("TotalPractices = %d, want 1", resp.Stats.TotalPractices)
	}

	// Progress fragment should reflect the saved practice.
	rec = do(e, http.MethodGet, "/progress", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /progress status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Progreso") {
		t.Error("GET /progress missing Progreso")
	}
}

func TestSubmitRejectsInvalidLevel(t *testing.T) {
	e := newTestServer(t)
	body := `{"level":99,"lang":"es","wordCount":20,"wpm":30,"accuracy":1,"consistency":10,"durationMS":1000}`
	rec := do(e, http.MethodPost, "/practice", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /practice invalid level status = %d, want 400", rec.Code)
	}
}

// Level 1 has no accented characters, so no dead-key hints are sent.
func TestPracticeLevel1HasNoAccentHints(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/practice?level=1&lang=es&count=20", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-accents=""`) {
		t.Errorf("level 1 should not expose accent hints, got: %s",
			firstMatch(rec.Body.String(), `data-accents="[^"]*"`))
	}
}

// Level 3 introduces acute accents, so the dead key and base letters are sent.
func TestPracticeLevel3AccentHints(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/practice?level=3&lang=es&count=20", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	raw := firstMatch(rec.Body.String(), `data-accents="([^"]*)"`)
	if raw == "" {
		t.Fatalf("missing data-accents in %s", rec.Body.String())
	}

	// The attribute carries escaped JSON (&#34; for quotes).
	encoded := strings.TrimSuffix(strings.TrimPrefix(raw, `data-accents="`), `"`)
	encoded = strings.ReplaceAll(encoded, "&#34;", `"`)
	encoded = strings.ReplaceAll(encoded, "&amp;", "&")

	var hints map[string]struct {
		Dead string `json:"dead"`
		Base string `json:"base"`
	}
	if err := json.Unmarshal([]byte(encoded), &hints); err != nil {
		t.Fatalf("decode accents %q: %v", encoded, err)
	}

	for _, tc := range []struct{ char, base string }{
		{"á", "a"}, {"é", "e"}, {"í", "i"}, {"ó", "o"}, {"ú", "u"},
	} {
		h, ok := hints[tc.char]
		if !ok {
			t.Errorf("missing hint for %q", tc.char)
			continue
		}
		if h.Dead != "´" {
			t.Errorf("%q dead key = %q, want ´", tc.char, h.Dead)
		}
		if h.Base != tc.base {
			t.Errorf("%q base = %q, want %q", tc.char, h.Base, tc.base)
		}
	}

	// Uppercase accents belong to level 4, not level 3.
	if _, ok := hints["Á"]; ok {
		t.Error("level 3 should not include Á")
	}
}

// Level 4 adds the uppercase accented vowels, which use SHIFT + base letter.
func TestPracticeLevel4UppercaseAccentHints(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/practice?level=4&lang=es&count=20", "")

	encoded := firstMatch(rec.Body.String(), `data-accents="([^"]*)"`)
	encoded = strings.ReplaceAll(encoded, "&#34;", `"`)

	var hints map[string]struct {
		Dead string `json:"dead"`
		Base string `json:"base"`
	}
	if err := json.Unmarshal([]byte(encoded), &hints); err != nil {
		t.Fatalf("decode accents: %v", err)
	}
	if h, ok := hints["Á"]; !ok || h.Base != "A" || h.Dead != "´" {
		t.Errorf("Á hint = %+v (present=%v), want dead ´ + base A", h, ok)
	}
}

// firstMatch returns the first regexp match (group 1 when present).
func firstMatch(s, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	if len(m) == 1 {
		return m[0]
	}
	return ""
}

// The progress fragment must carry the chart series, oldest first, so the
// client can plot WPM and accuracy over time (S4).
func TestProgressIncludesChartSeries(t *testing.T) {
	e := newTestServer(t)

	for _, wpm := range []string{"20.0", "30.0", "40.0"} {
		body := `{"level":1,"lang":"es","wordCount":20,"wpm":` + wpm +
			`,"accuracy":0.9,"consistency":30,"durationMS":60000}`
		if rec := do(e, http.MethodPost, "/practice", body); rec.Code != http.StatusOK {
			t.Fatalf("POST /practice status = %d", rec.Code)
		}
	}

	rec := do(e, http.MethodGet, "/progress", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /progress status = %d", rec.Code)
	}

	encoded := firstMatch(rec.Body.String(), `data-series="([^"]*)"`)
	if encoded == "" {
		t.Fatalf("missing data-series in %s", rec.Body.String())
	}
	encoded = strings.ReplaceAll(encoded, "&#34;", `"`)
	encoded = strings.ReplaceAll(encoded, "&amp;", "&")

	var points []struct {
		Label string  `json:"label"`
		WPM   float64 `json:"wpm"`
		Acc   float64 `json:"acc"`
	}
	if err := json.Unmarshal([]byte(encoded), &points); err != nil {
		t.Fatalf("decode series %q: %v", encoded, err)
	}

	if len(points) != 3 {
		t.Fatalf("got %d points, want 3", len(points))
	}
	// Oldest first: 20, 30, 40.
	wantWPM := []float64{20, 30, 40}
	for i, want := range wantWPM {
		if points[i].WPM != want {
			t.Errorf("point %d WPM = %v, want %v", i, points[i].WPM, want)
		}
		if points[i].Acc != 90 {
			t.Errorf("point %d acc = %v, want 90", i, points[i].Acc)
		}
		if points[i].Label == "" {
			t.Errorf("point %d has an empty label", i)
		}
	}
}

// Without practices the chart has an empty series and still renders.
func TestProgressEmptySeries(t *testing.T) {
	e := newTestServer(t)
	rec := do(e, http.MethodGet, "/progress", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-series="[]"`) {
		t.Errorf("expected empty series, got %s",
			firstMatch(rec.Body.String(), `data-series="[^"]*"`))
	}
}
