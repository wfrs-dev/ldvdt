package static

import (
	"io/fs"
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
