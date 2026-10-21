// Package sqlite implements the storage port over a local SQLite database.
// It uses modernc.org/sqlite (pure Go, no CGO) so the binary cross-compiles
// to Windows and Linux.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS practices (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	language     TEXT    NOT NULL,
	level        INTEGER NOT NULL,
	wpm          REAL    NOT NULL,
	accuracy     REAL    NOT NULL,
	consistency  REAL    NOT NULL,
	duration_ms  INTEGER NOT NULL,
	word_count   INTEGER NOT NULL,
	completed_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_practices_language ON practices(language);
`

// Store is a SQLite-backed Storage implementation.
type Store struct {
	db *sql.DB
}

// Open opens (and creates if needed) the SQLite database at path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}
	// A single local writer; serialize access to avoid SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// SavePractice inserts a completed practice and returns its id.
func (s *Store) SavePractice(ctx context.Context, p domain.Practice) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO practices (language, level, wpm, accuracy, consistency, duration_ms, word_count, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		string(p.Language), p.Level, p.WPM, p.Accuracy, p.Consistency,
		p.DurationMS, p.WordCount, p.CompletedAt.UnixMilli(),
	)
	if err != nil {
		return 0, fmt.Errorf("sqlite: save practice: %w", err)
	}
	return res.LastInsertId()
}

// Stats returns aggregated stats for a language ("" means global).
func (s *Store) Stats(ctx context.Context, language domain.Language) (domain.Stats, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*),
		        COALESCE(SUM(duration_ms), 0),
		        COALESCE(MAX(wpm), 0),
		        COALESCE(AVG(wpm), 0),
		        COALESCE(AVG(accuracy), 0),
		        COALESCE(AVG(consistency), 0)
		   FROM practices
		  WHERE (? = '' OR language = ?)`,
		string(language), string(language),
	)

	var st domain.Stats
	if err := row.Scan(&st.TotalPractices, &st.TotalTimeMS, &st.BestWPM,
		&st.AvgWPM, &st.AvgAccuracy, &st.AvgConsistency); err != nil {
		return domain.Stats{}, fmt.Errorf("sqlite: stats: %w", err)
	}
	return st, nil
}

// RecentPractices returns the most recent practices, newest first.
func (s *Store) RecentPractices(ctx context.Context, language domain.Language, limit int) ([]domain.Practice, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, language, level, wpm, accuracy, consistency, duration_ms, word_count, completed_at
		   FROM practices
		  WHERE (? = '' OR language = ?)
		  ORDER BY completed_at DESC, id DESC
		  LIMIT ?`,
		string(language), string(language), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: recent practices: %w", err)
	}
	defer rows.Close()

	var out []domain.Practice
	for rows.Next() {
		var p domain.Practice
		var lang string
		var completedAt int64
		if err := rows.Scan(&p.ID, &lang, &p.Level, &p.WPM, &p.Accuracy,
			&p.Consistency, &p.DurationMS, &p.WordCount, &completedAt); err != nil {
			return nil, fmt.Errorf("sqlite: scan practice: %w", err)
		}
		p.Language = domain.Language(lang)
		p.CompletedAt = time.UnixMilli(completedAt)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate practices: %w", err)
	}
	return out, nil
}
