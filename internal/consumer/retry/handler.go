package retry

import (
	"context"
	"errors"
	"time"

	"github.com/gee-kul/orderflow/internal/event"
)

type EventHandler interface {
	Handle(ctx context.Context, evt event.Event) error
}

type Handler struct {
	next           EventHandler
	maxAttempts    int
	initialBackoff time.Duration
	maxBackoff     time.Duration
}

var (
	ErrHandlerRequired       = errors.New("handler is required")
	ErrMaxAttemptsInvalid    = errors.New("max attempts invalid")
	ErrInitialBackoffInvalid = errors.New("initial backoff invalid")
	ErrMaxBackoffInvalid     = errors.New("max backoff invalid")
)

func NewHandler(next EventHandler, maxAttempts int,
	initialBackoff time.Duration, maxBackoff time.Duration) (*Handler, error) {
	if next == nil {
		return nil, ErrHandlerRequired
	}
	if maxAttempts < 1 {
		return nil, ErrMaxAttemptsInvalid
	}
	if initialBackoff < 0 {
		return nil, ErrInitialBackoffInvalid
	}
	if maxBackoff < initialBackoff {
		return nil, ErrMaxBackoffInvalid
	}

	handler := Handler{next: next, maxAttempts: maxAttempts,
		initialBackoff: initialBackoff, maxBackoff: maxBackoff}

	return &handler, nil
}

func (h *Handler) Handle(ctx context.Context, evt event.Event) error {
	var lastErr error

	backoff := h.initialBackoff

	for attempt := range h.maxAttempts {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = h.next.Handle(ctx, evt)
		if lastErr == nil {
			return nil
		}

		if attempt == h.maxAttempts-1 {
			return lastErr
		}

		timer := time.NewTimer(backoff)

		select {
		case <-timer.C:

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}

		backoff = nextBackoff(backoff, h.maxBackoff)

	}
	return lastErr
}

func nextBackoff(current time.Duration, maximum time.Duration) time.Duration {
	if current > maximum/2 {
		return maximum
	}
	return current * 2
}
