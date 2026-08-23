package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	ErrBrokersRequired = errors.New("brokers required")
	ErrTopicRequired   = errors.New("topic required")
)

type message struct {
	EventID       string          `json:"event_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Publisher struct {
	client *kgo.Client
	topic  string
}

func NewPublisher(brokers []string, topic string) (*Publisher, error) {
	if len(brokers) == 0 {
		return nil, ErrBrokersRequired
	}

	if topic == "" {
		return nil, ErrTopicRequired
	}

	opt := kgo.SeedBrokers(brokers...)
	prodOpt := kgo.RequiredAcks(kgo.AllISRAcks())
	prodOpt2 := kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil))

	client, err := kgo.NewClient(opt, prodOpt, prodOpt2)
	if err != nil {
		return nil, err
	}

	var publisher Publisher
	publisher.client = client
	publisher.topic = topic

	return &publisher, nil
}

func (p *Publisher) Publish(ctx context.Context, evt event.Event) error {
	msg := message{EventID: evt.ID, AggregateType: evt.AggregateType,
		AggregateID: evt.AggregateID, EventType: evt.EventType, Payload: evt.Payload,
		CreatedAt: evt.CreatedAt}

	messageFromMarshal, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	record := kgo.Record{Topic: p.topic, Key: []byte(evt.AggregateID), Value: messageFromMarshal}

	res := p.client.ProduceSync(ctx, &record)

	return res.FirstErr()
}

func (p *Publisher) Close() {
	p.client.Close()
}
