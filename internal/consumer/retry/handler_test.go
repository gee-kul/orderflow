package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gee-kul/orderflow/internal/event"
)

type fakeHandler struct {
	calls    int
	failures int
	err      error
	onCall   func()
}

func (h *fakeHandler) Handle(ctx context.Context, evt event.Event) error {
	h.calls++
	if h.onCall != nil {
		h.onCall()
	}
	if h.calls <= h.failures {
		return h.err
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name           string
		next           EventHandler
		maxAttempts    int
		initialBackoff time.Duration
		maxBackoff     time.Duration
		wantErr        error
	}{
		{
			name:           "no handler",
			next:           nil,
			maxAttempts:    1,
			initialBackoff: time.Second,
			maxBackoff:     2 * time.Second,
			wantErr:        ErrHandlerRequired,
		},
		{
			name:           "zero attempts",
			next:           &fakeHandler{},
			maxAttempts:    0,
			initialBackoff: time.Second,
			maxBackoff:     2 * time.Second,
			wantErr:        ErrMaxAttemptsInvalid,
		},
		{
			name:           "negative initialBackoff",
			next:           &fakeHandler{},
			maxAttempts:    1,
			initialBackoff: -time.Second,
			maxBackoff:     2 * time.Second,
			wantErr:        ErrInitialBackoffInvalid,
		},
		{
			name:           "maxBackoff < initialBackoff",
			next:           &fakeHandler{},
			maxAttempts:    1,
			initialBackoff: 2 * time.Second,
			maxBackoff:     time.Second,
			wantErr:        ErrMaxBackoffInvalid,
		},
		{
			name:           "success",
			next:           &fakeHandler{},
			maxAttempts:    1,
			initialBackoff: time.Second,
			maxBackoff:     2 * time.Second,
			wantErr:        nil,
		}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(test.next, test.maxAttempts, test.initialBackoff, test.maxBackoff)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ошибка должна быть: %v, а вывелось: %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				if handler != nil {
					t.Error("handler must be nil")
				}
				return
			}
			if handler == nil {
				t.Fatal("handler must not be nil")
			}

			if handler.next != test.next {
				t.Errorf("next must be %v, but received %v", test.next, handler.next)
			}
			if handler.maxAttempts != test.maxAttempts {
				t.Errorf("maxAttempts must be %v, but received %v", test.maxAttempts, handler.maxAttempts)
			}
			if handler.initialBackoff != test.initialBackoff {
				t.Errorf("initialBackoff must be %v, but received %v", test.initialBackoff, handler.initialBackoff)
			}
			if handler.maxBackoff != test.maxBackoff {
				t.Errorf("maxBackoff must be %v, but received %v", test.maxBackoff, handler.maxBackoff)
			}

		})
	}
}

func TestHandlerHandle(t *testing.T) {
	handleErr := errors.New("handler failed")
	tests := []struct {
		name        string
		failures    int
		maxAttempts int
		wantCalls   int
		wantErr     error
	}{
		{
			name:        "success immediately",
			failures:    0,
			maxAttempts: 3,
			wantCalls:   1,
			wantErr:     nil,
		},
		{
			name:        "success from 3 attempt",
			failures:    2,
			maxAttempts: 3,
			wantCalls:   3,
			wantErr:     nil,
		},
		{
			name:        "no success",
			failures:    3,
			maxAttempts: 3,
			wantCalls:   3,
			wantErr:     handleErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeHand := fakeHandler{failures: test.failures, err: handleErr}
			handler, err := NewHandler(&fakeHand, test.maxAttempts, 0, 0)
			if err != nil {
				t.Fatalf("ошибка должна быть nil, а вывелось: %v", err)
			}

			err = handler.Handle(t.Context(), event.Event{})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error must be %v, but received %v", test.wantErr, err)
			}

			if fakeHand.calls != test.wantCalls {
				t.Errorf("calls must be %v, but received %v", test.wantCalls, fakeHand.calls)
			}

		})
	}

}

func TestNextBackoff(t *testing.T) {
	tests := []struct {
		name    string
		current time.Duration
		maximum time.Duration
		want    time.Duration
	}{
		{
			name:    "first case",
			current: time.Second,
			maximum: 5 * time.Second,
			want:    2 * time.Second,
		},
		{
			name:    "second case",
			current: 4 * time.Second,
			maximum: 5 * time.Second,
			want:    5 * time.Second,
		},
		{
			name:    "third case",
			current: 5 * time.Second,
			maximum: 5 * time.Second,
			want:    5 * time.Second,
		},
		{
			name:    "fourth case",
			current: 0 * time.Second,
			maximum: 5 * time.Second,
			want:    0 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backoff := nextBackoff(test.current, test.maximum)

			if backoff != test.want {
				t.Errorf("want duration must be %v, but received %v", test.want, backoff)
			}

		})
	}
}

func TestHandlerHandleCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	fake := fakeHandler{}
	handler, err := NewHandler(&fake, 3, 0, 0)
	if err != nil {
		t.Fatalf("ошибка должна быть nil, а вывелось: %v", err)
	}

	err = handler.Handle(ctx, event.Event{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error must be %v, but received %v", context.Canceled, err)
	}

	if fake.calls != 0 {
		t.Errorf("calls must be 0, but received %v", fake.calls)
	}
}

func TestHandlerHandleCancellationStopsRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	fake := fakeHandler{failures: 1, err: errors.New("error"), onCall: cancel}

	handler, err := NewHandler(&fake, 3, time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("ошибка должна быть nil, а вывелось: %v", err)
	}

	ch := make(chan error, 1)

	go func() {
		err := handler.Handle(ctx, event.Event{})
		ch <- err
	}()

	select {
	case recerr := <-ch:
		if !errors.Is(recerr, context.Canceled) {
			t.Fatalf("error must be %v, but received %v", context.Canceled, recerr)
		}
	case <-time.After(time.Second):
		t.Fatal("handler doesnt stop")
	}

	if fake.calls != 1 {
		t.Errorf("calls must be 1, but received %v", fake.calls)
	}

}
