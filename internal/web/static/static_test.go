package static

import (
	"io/fs"
	"strings"
	"testing"
)

// TestEmbeddedFonts pins that the only embedded fonts are the two Victor Mono
// woff2 files the UI actually loads (S19). Every other weight/style was removed
// so the binary does not ship dead font assets.
func TestEmbeddedFonts(t *testing.T) {
	matches, err := fs.Glob(FS, "fonts/*")
	if err != nil {
		t.Fatalf("glob fonts: %v", err)
	}

	want := map[string]bool{
		"fonts/VictorMono-Regular.woff2": true,
		"fonts/VictorMono-Italic.woff2":  true,
	}

	if len(matches) != len(want) {
		t.Fatalf("embedded fonts = %v, want exactly %d files", matches, len(want))
	}
	for _, m := range matches {
		if !want[m] {
			t.Errorf("unexpected embedded font %q", m)
		}
	}
}

// TestVendoredBootstrapTheme pins the vendored Bootstrap build (S21). The file
// is compiled from docs/bootstrap-theme.scss with the palette the user asked
// for, so the theme has to reach Bootstrap's own custom properties — patching
// the stock build would leave buttons and the form focus ring on the default
// colours. Normalising whitespace keeps the assertions independent of the Sass
// output style (expanded vs compressed).
func TestVendoredBootstrapTheme(t *testing.T) {
	raw, err := fs.ReadFile(FS, "vendor/bootstrap.min.css")
	if err != nil {
		t.Fatalf("read vendored bootstrap: %v", err)
	}
	css := stripSpaces(string(raw))

	// Every colour of the requested palette, as Bootstrap's own tokens.
	palette := []string{
		"--bs-primary:#26a269",
		"--bs-primary-rgb:38,162,105",
		"--bs-secondary:#6c757d",
		"--bs-success:#198754",
		"--bs-info:#0dcaf0",
		"--bs-warning:#ffc107",
		"--bs-danger:#dc3545",
		"--bs-light:#f8f9fa",
		"--bs-dark:#212529",
	}
	for _, want := range palette {
		if !strings.Contains(css, want) {
			t.Errorf("vendored bootstrap is missing the themed token %q", want)
		}
	}

	// The primary colour also has to reach the component layer: Bootstrap 5.3
	// hardcodes it in .btn-primary, which is exactly what a CSS-only override
	// of --bs-primary fails to recolour.
	if !strings.Contains(css, "--bs-btn-bg:#26a269") {
		t.Error("vendored bootstrap does not theme .btn-primary: expected --bs-btn-bg:#26a269")
	}

	// Dark mode must come from Bootstrap's own attribute (S24).
	if !strings.Contains(css, "[data-bs-theme=dark]{") {
		t.Error("vendored bootstrap has no [data-bs-theme=dark] block")
	}

	// MIT header stays with the vendored code.
	for _, want := range []string{"BootstrapAuthors", "LicensedunderMIT"} {
		if !strings.Contains(css, want) {
			t.Errorf("vendored bootstrap is missing %q", want)
		}
	}
}

// stripSpaces removes every whitespace run so CSS assertions do not depend on
// the Sass output style.
func stripSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}
