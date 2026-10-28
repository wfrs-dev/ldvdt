package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

// SpaceKeyID is the SVG id of the spacebar. The spacebar is not part of
// layout.yaml, which describes the 48 character keys, so its id is named here.
const SpaceKeyID = "space"

// keymapJSON returns the character -> SVG element id map the client engine uses
// to repaint the key to press (S23). Both shift states of a key resolve to the
// same element: the SVG key is one rect regardless of SHIFT. The spacebar has no
// layout key and maps to the design's own #space.
func keymapJSON(kb domain.Keyboard) string {
	keys := kb.Keys()
	keymap := make(map[string]string, 2*len(keys)+1)
	for _, k := range keys {
		id := "key" + strconv.Itoa(k.ID)
		keymap[string(k.Top)] = id
		keymap[string(k.Bottom)] = id
	}
	keymap[" "] = SpaceKeyID

	encoded, err := json.Marshal(keymap)
	if err != nil {
		// A map[string]string cannot fail to marshal; the empty map keeps the
		// page usable instead of rendering a broken attribute.
		return "{}"
	}
	return string(encoded)
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

// progressPoint is one plotted practice: a label plus WPM and accuracy percent.
type progressPoint struct {
	Label string  `json:"label"`
	WPM   float64 `json:"wpm"`
	Acc   float64 `json:"acc"`
}

// chartLabel renders a practice timestamp for the chart axis in now's location.
// Today's practices show the 12-hour time only; older ones also carry the full
// date (day/month/year), so points from different days and years stay
// distinguishable. at is normalised to now's location so the "today" boundary
// and the rendered clock agree even when the two times carry different zones.
func chartLabel(at, now time.Time) string {
	at = at.In(now.Location())
	if at.Year() == now.Year() && at.YearDay() == now.YearDay() {
		return at.Format("03:04PM")
	}
	return at.Format("02/01/2006 03:04PM")
}

// SeriesJSON serialises the practices to the compact JSON the client chart
// reads from the canvas data attribute (S4). Practices must be oldest first and
// now anchors the labels. A serialisation failure is returned to the caller
// instead of being hidden behind an empty chart.
func SeriesJSON(practices []domain.Practice, now time.Time) (string, error) {
	if len(practices) == 0 {
		return "[]", nil
	}

	points := make([]progressPoint, len(practices))
	for i, p := range practices {
		points[i] = progressPoint{
			Label: chartLabel(p.CompletedAt, now),
			WPM:   math.Round(p.WPM*10) / 10,
			Acc:   math.Round(p.Accuracy*1000) / 10,
		}
	}

	data, err := json.Marshal(points)
	if err != nil {
		return "", fmt.Errorf("encode progress series: %w", err)
	}
	return string(data), nil
}
