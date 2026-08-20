package outbox

import (
	"context"

	"github.com/gee-kul/orderflow/internal/event"
)

type Publisher interface {
	Publish(ctx context.Context, evt event.Event) error
}