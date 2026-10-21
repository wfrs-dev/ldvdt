// Package trainer implements the trainer module: use cases, HTTP handlers and
// routing (hexagonal: handler -> application -> port/storage).
package trainer

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/wfrs-dev/ldvdt/internal/corpus"
	"github.com/wfrs-dev/ldvdt/internal/domain"
	"github.com/wfrs-dev/ldvdt/internal/port"
)

const (
	minWordCount = 20
	maxWordCount = 500
)

// Service is the application layer for the trainer use cases.
type Service struct {
	store  port.Storage
	corpus *corpus.Corpus
	levels []domain.Level
	rng    *rand.Rand
}

// NewService wires the application dependencies.
func NewService(store port.Storage, c *corpus.Corpus, levels []domain.Level) *Service {
	return &Service{
		store:  store,
		corpus: c,
		levels: levels,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Levels returns the practice levels.
func (s *Service) Levels() []domain.Level { return s.levels }

// Practice selects the words for a practice session.
func (s *Service) Practice(levelNum int, lang domain.Language, count int) ([]string, domain.Level, error) {
	level, err := s.level(levelNum)
	if err != nil {
		return nil, domain.Level{}, err
	}
	if !lang.Valid() {
		return nil, domain.Level{}, fmt.Errorf("idioma no válido: %q", lang)
	}
	if count < minWordCount {
		count = minWordCount
	}
	if count > maxWordCount {
		count = maxWordCount
	}
	return s.corpus.Select(level, lang, count, s.rng), level, nil
}

// SaveResult persists a completed practice and returns the updated global stats.
func (s *Service) SaveResult(ctx context.Context, r SubmitRequest) (domain.Stats, error) {
	if r.Level < 1 || r.Level > len(s.levels) {
		return domain.Stats{}, fmt.Errorf("nivel no válido: %d", r.Level)
	}
	lang := domain.Language(r.Lang)
	if !lang.Valid() {
		return domain.Stats{}, fmt.Errorf("idioma no válido: %q", r.Lang)
	}

	p := domain.Practice{
		Language:    lang,
		Level:       r.Level,
		WPM:         r.WPM,
		Accuracy:    r.Accuracy,
		Consistency: r.Consistency,
		DurationMS:  r.DurationMS,
		WordCount:   r.WordCount,
		CompletedAt: time.Now(),
	}
	if _, err := s.store.SavePractice(ctx, p); err != nil {
		return domain.Stats{}, err
	}
	return s.store.Stats(ctx, "")
}

// Progress returns aggregate stats for a language ("" means global).
func (s *Service) Progress(ctx context.Context, lang domain.Language) (domain.Stats, error) {
	if lang != "" && !lang.Valid() {
		return domain.Stats{}, fmt.Errorf("idioma no válido: %q", lang)
	}
	return s.store.Stats(ctx, lang)
}

func (s *Service) level(n int) (domain.Level, error) {
	if n < 1 || n > len(s.levels) {
		return domain.Level{}, fmt.Errorf("nivel no válido: %d", n)
	}
	return s.levels[n-1], nil
}
