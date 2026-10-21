package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSaveAndStats(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	practices := []domain.Practice{
		{Language: domain.LanguageSpanish, Level: 1, WPM: 30, Accuracy: 0.95, Consistency: 40, DurationMS: 60000, WordCount: 20, CompletedAt: now.Add(-2 * time.Minute)},
		{Language: domain.LanguageSpanish, Level: 2, WPM: 34, Accuracy: 0.97, Consistency: 35, DurationMS: 60000, WordCount: 20, CompletedAt: now.Add(-1 * time.Minute)},
		{Language: domain.LanguageEnglish, Level: 1, WPM: 40, Accuracy: 0.90, Consistency: 50, DurationMS: 30000, WordCount: 20, CompletedAt: now},
	}

	for i := range practices {
		id, err := s.SavePractice(ctx, practices[i])
		if err != nil {
			t.Fatalf("SavePractice %d: %v", i, err)
		}
		if id == 0 {
			t.Fatalf("SavePractice %d returned id 0", i)
		}
	}

	global, err := s.Stats(ctx, "")
	if err != nil {
		t.Fatalf("Stats global: %v", err)
	}
	if global.TotalPractices != 3 {
		t.Errorf("global practices = %d, want 3", global.TotalPractices)
	}
	if global.TotalTimeMS != 150000 {
		t.Errorf("global total time = %d, want 150000", global.TotalTimeMS)
	}
	if global.BestWPM != 40 {
		t.Errorf("global best WPM = %v, want 40", global.BestWPM)
	}

	es, err := s.Stats(ctx, domain.LanguageSpanish)
	if err != nil {
		t.Fatalf("Stats es: %v", err)
	}
	if es.TotalPractices != 2 {
		t.Errorf("es practices = %d, want 2", es.TotalPractices)
	}
	if es.TotalTimeMS != 120000 {
		t.Errorf("es total time = %d, want 120000", es.TotalTimeMS)
	}
	if es.BestWPM != 34 {
		t.Errorf("es best WPM = %v, want 34", es.BestWPM)
	}
}

func TestRecentPracticesOrdering(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 5; i++ {
		if _, err := s.SavePractice(ctx, domain.Practice{
			Language: domain.LanguageEnglish, Level: 1, WPM: float64(i),
			CompletedAt: now.Add(time.Duration(i) * time.Minute),
		}); err != nil {
			t.Fatalf("SavePractice: %v", err)
		}
	}

	recent, err := s.RecentPractices(ctx, "", 3)
	if err != nil {
		t.Fatalf("RecentPractices: %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("got %d, want 3", len(recent))
	}
	// Newest first.
	if recent[0].WPM != 4 || recent[1].WPM != 3 || recent[2].WPM != 2 {
		t.Errorf("unexpected order: %v %v %v", recent[0].WPM, recent[1].WPM, recent[2].WPM)
	}
}
