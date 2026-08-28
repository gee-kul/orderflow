package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type fakeHandler struct {
	evts    []event.Event
	wantErr error
}

func (h *fakeHandler) Handle(ctx context.Context, evt event.Event) error {
	h.evts = append(h.evts, evt)
	return h.wantErr
}

func TestNewConsumer(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		topic   string
		group   string
		handler Handler
		wantErr error
	}{
		{
			name:    "noBrokers",
			brokers: nil,
			topic:   "order.events",
			group:   "orderflow-order-stats",
			handler: &fakeHandler{},
			wantErr: ErrBrokersRequired,
		},
		{
			name:    "noTopic",
			brokers: []string{"localhost:9092"},
			topic:   "",
			group:   "orderflow-order-stats",
			handler: &fakeHandler{},
			wantErr: ErrTopicRequired,
		},
		{
			name:    "noGroup",
			brokers: []string{"localhost:9092"},
			topic:   "order.events",
			group:   "",
			handler: &fakeHandler{},
			wantErr: ErrGroupRequired,
		},
		{
			name:    "noHandler",
			brokers: []string{"localhost:9092"},
			topic:   "order.events",
			group:   "orderflow-order-stats",
			handler: nil,
			wantErr: ErrHandlerRequired,
		},
		{
			name:    "success",
			brokers: []string{"localhost:9092"},
			topic:   "order.events",
			group:   "orderflow-order-stats",
			handler: &fakeHandler{},
			wantErr: nil,
		}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumer, err := NewConsumer(test.brokers, test.topic, test.group, test.handler)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ошибка должна быть: %v, а вывелось: %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				if consumer != nil {
					t.Error("consumer must be nil")
				}
				return
			}
			if consumer == nil {
				t.Fatal("consumer must not be nil")
			}

			if consumer.client == nil {
				t.Fatal("consumer.client must be not nil")
			}
			defer consumer.Close()
			if consumer.handler != test.handler {
				t.Errorf("handler must be: %v, but received: %v", test.handler, consumer.handler)
			}

		})
	}
}

func TestConsumerProcessRecord(t *testing.T) {
	evt := event.Event{ID: uuid.NewString()}
	envelope := event.NewEnvelope(evt)
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("error from matshal: %v", err)
	}

	var record kgo.Record
	record.Value = envelopeJSON

	var handler fakeHandler

	consumer := Consumer{handler: &handler}
	err = consumer.processRecord(t.Context(), &record)
	if err != nil {
		t.Errorf("error from processRecord: %v", err)
	}
	if len(handler.evts) != 1 {
		t.Errorf("len handler must be 1, but received %v", len(handler.evts))
	}
}

func TestConsumerProcessRecordInvalidJSON(t *testing.T) {
	record := &kgo.Record{
		Value: []byte(`not-json`),
	}

	handler := &fakeHandler{}
	consumer := Consumer{handler: handler}

	err := consumer.processRecord(t.Context(), record)
	if err == nil {
		t.Fatal("error must be not nil")
	}
	if len(handler.evts) != 0 {
		t.Errorf("len handler must be 0, but received %v", len(handler.evts))
	}
}

func TestConsumerProcessRecordHandlerError(t *testing.T) {
	evt := event.Event{ID: uuid.NewString()}
	envelope := event.NewEnvelope(evt)
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("error from matshal: %v", err)
	}

	var record kgo.Record
	record.Value = envelopeJSON

	handlerErr := errors.New("handler failed")
	handler := fakeHandler{wantErr: handlerErr}
	consumer := Consumer{handler: &handler}

	err = consumer.processRecord(t.Context(), &record)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error, received: %v", err)
	}
	if len(handler.evts) != 1 {
		t.Errorf("len handler must be 1, but received %v", len(handler.evts))
	}
}
