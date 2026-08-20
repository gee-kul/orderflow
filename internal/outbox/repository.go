package outbox

import (
	"context"

	"github.com/gee-kul/orderflow/internal/event"
)

type Repository interface {
	FetchUnpublished(ctx context.Context, limit int) ([]event.Event, error)
	MarkPublished(ctx context.Context, id string) error
}