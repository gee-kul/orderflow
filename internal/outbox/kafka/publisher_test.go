package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestNewPublisher(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		topic   string
		wantErr error
	}{
		{
			name:    "noBrokers",
			brokers: nil,
			topic:   "order.events",
			wantErr: ErrBrokersRequired,
		},
		{
			name:    "noTopic",
			brokers: []string{"localhost:9092"},
			topic:   "",
			wantErr: ErrTopicRequired,
		},
		{
			name:    "success",
			brokers: []string{"localhost:9092"},
			topic:   "order.events",
			wantErr: nil,
		}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher, err := NewPublisher(test.brokers, test.topic)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ошибка должна быть: %v, а вывелось: %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				if publisher != nil {
					t.Error("publisher must be nil")
				}
				return
			}
			if publisher == nil {
				t.Fatal("publisher must not be nil")
			}

			defer publisher.Close()

			if publisher.client == nil {
				t.Error("publisher.client must be not nil")
			}
			if publisher.topic != test.topic {
				t.Errorf("topic must be: %v, but replay: %v", test.topic, publisher.topic)
			}

		})
	}
}

func TestPublisherPublishMarshalError(t *testing.T) {
	evt := event.Event{ID: "id-1", Payload: json.RawMessage{}}
	publisher := Publisher{}
	err := publisher.Publish(t.Context(), evt)
	if err == nil {
		t.Fatal("error from publish must not be nil")
	}
}

func TestPublisherPublishIntegration(t *testing.T) {
	str := os.Getenv("ORDERFLOW_TEST_KAFKA_BROKERS")
	if str == "" {
		t.Skip("ORDERFLOW_TEST_KAFKA_BROKERS не задана")
	}
	brokers := strings.Split(str, ",")

	publisher, err := NewPublisher(brokers, "order.events")
	if err != nil {
		t.Fatalf("error from NewPublisher %v", err)
	}

	defer publisher.Close()

	evt := event.Event{ID: uuid.NewString(), AggregateType: "order", AggregateID: uuid.NewString(),
		EventType: "order.created", Payload: json.RawMessage(`{"customer_id":"customer-1","status":"created"}`), CreatedAt: time.Date(2026, 8, 20, 17, 22, 0, 0, time.UTC)}

	ctx := t.Context()
	ctxNew, canc := context.WithTimeout(ctx, 10*time.Second)
	defer canc()
	err = publisher.Publish(ctxNew, evt)
	if err != nil {
		t.Fatalf("error from publish must be nil, not %v", err)
	}

	opt1 := kgo.SeedBrokers(brokers...)
	opt2 := kgo.ConsumeTopics("order.events")
	opt3 := kgo.ConsumeResetOffset(kgo.NewOffset().AtStart())
	consumer, err := kgo.NewClient(opt1, opt2, opt3)
	if err != nil {
		t.Fatalf("error from NewClient %v", err)
	}

	defer consumer.Close()

	for {
		fetches := consumer.PollFetches(ctxNew)
		if ctxNew.Err() != nil {
			t.Fatalf("context is done %v", ctxNew.Err())
		}

		if len(fetches.Errors()) != 0 {
			t.Fatalf("error from fetches %v", fetches.Errors())
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			if string(record.Key) == evt.AggregateID {
				var message event.Envelope
				err = json.Unmarshal(record.Value, &message)
				if err != nil {
					t.Fatalf("error from unmarshal %v", err)
				}

				if message.EventID != evt.ID {
					t.Errorf("event_id didnt metch, ожидали %v", evt.ID)
				}
				if message.AggregateType != evt.AggregateType {
					t.Errorf("aggregate_type didnt metch, ожидали %v", evt.AggregateType)
				}
				if message.AggregateID != evt.AggregateID {
					t.Errorf("aggregate_id didnt metch, ожидали %v", evt.AggregateID)
				}
				if message.EventType != evt.EventType {
					t.Errorf("event_type didnt metch, ожидали %v", evt.EventType)
				}
				if string(message.Payload) != string(evt.Payload) {
					t.Errorf("payload didnt metch, ожидали %v", string(evt.Payload))
				}
				if !message.CreatedAt.Equal(evt.CreatedAt) {
					t.Errorf("created_at didnt metch, ожидали %v", evt.CreatedAt)
				}
				if record.Topic != "order.events" {
					t.Errorf("topic didnt metch, ожидали order.events")
				}
				s := make(map[string]json.RawMessage)
				err = json.Unmarshal(record.Value, &s)
				_, ok := s["published_at"]
				if ok {
					t.Fatal("published_at must not exists")
				}
				return

			}
		}
	}
}
