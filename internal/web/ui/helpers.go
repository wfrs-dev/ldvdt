package ui

import (
	"fmt"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

// keyClass returns the CSS class for a key, including its finger color group.
func keyClass(k domain.Key) string {
	return "key finger-" + string(k.Finger())
}

// formatDuration renders milliseconds as a compact human duration.
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	total := ms / 1000
	m := total / 60
	s := total % 60
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// pct renders a 0..1 fraction as a percentage with one decimal.
func pct(f float64) string {
	return fmt.Sprintf("%.1f%%", f*100)
}

// f1 renders a float with one decimal.
func f1(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
