// Package port declares the outbound interfaces (ports) that the application
// layer depends on, following hexagonal architecture.
package port

import (
	"context"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

// Storage is the persistence port for practice progress.
type Storage interface {
	// SavePractice persists a completed practice and returns its id.
	SavePractice(ctx context.Context, p domain.Practice) (int64, error)
	// Stats returns aggregated stats. A language with value "" means global.
	Stats(ctx context.Context, language domain.Language) (domain.Stats, error)
	// RecentPractices returns the most recent practices, newest first.
	RecentPractices(ctx context.Context, language domain.Language, limit int) ([]domain.Practice, error)
	// Close releases the underlying resources.
	Close() error
}
