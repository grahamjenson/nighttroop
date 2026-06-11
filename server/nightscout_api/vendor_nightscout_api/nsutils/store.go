package nsutils

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrGone     = errors.New("record deleted")
)

// Store abstracts persistence for Nightscout collections.
//
// Applications are expected to provide their own implementation.
type Store interface {
	Create(ctx context.Context, collection Collection, data map[string]any, subject string) (Record, bool, error)
	Get(ctx context.Context, collection Collection, identifier string) (Record, error)
	Search(ctx context.Context, collection Collection, query Query) ([]Record, error)
	Update(ctx context.Context, collection Collection, identifier string, patch map[string]any, subject string) (Record, error)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool, subject string) error
	DeleteMatching(ctx context.Context, collection Collection, query Query, permanent bool, subject string) (int, error)
}
