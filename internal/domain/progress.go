package domain

import "time"

// Practice is one completed typing practice, as persisted (S11).
type Practice struct {
	ID          int64
	Language    Language
	Level       int
	WPM         float64
	Accuracy    float64 // 0..1 (fraction of correct keystrokes)
	Consistency float64 // regularity of keystroke rhythm (ms standard deviation)
	DurationMS  int64
	WordCount   int
	CompletedAt time.Time
}

// Stats aggregates practice progress (S11).
type Stats struct {
	TotalPractices int
	TotalTimeMS    int64
	BestWPM        float64
	AvgWPM         float64
	AvgAccuracy    float64
	AvgConsistency float64
}
