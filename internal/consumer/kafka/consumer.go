package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Handler interface {
	Handle(ctx context.Context, evt event.Event) error
}

type Consumer struct {
	client  *kgo.Client
	handler Handler
}

var (
	ErrBrokersRequired = errors.New("brokers are required")
	ErrTopicRequired   = errors.New("topic is required")
	ErrGroupRequired   = errors.New("group is required")
	ErrHandlerRequired = errors.New("handler is required")
)

func NewConsumer(brokers []string, topic string, group string, handler Handler) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, ErrBrokersRequired
	}
	if topic == "" {
		return nil, ErrTopicRequired
	}
	if group == "" {
		return nil, ErrGroupRequired
	}
	if handler == nil {
		return nil, ErrHandlerRequired
	}
	opt := kgo.SeedBrokers(brokers...)
	opt2 := kgo.ConsumeTopics(topic)
	opt3 := kgo.ConsumerGroup(group)
	opt4 := kgo.DisableAutoCommit()
	client, err := kgo.NewClient(opt, opt2, opt3, opt4)
	if err != nil {
		return nil, err
	}

	consumer := Consumer{client: client, handler: handler}
	return &consumer, nil
}

func (con *Consumer) Close() {
	con.client.Close()
}

func (con *Consumer) Run(ctx context.Context) error {
	for {
		fetches := con.client.PollRecords(ctx, 1)
		if ctx.Err() != nil {
			return nil
		}
		errorsFromFetch := fetches.Errors()
		if len(errorsFromFetch) != 0 {
			return fmt.Errorf("error from errors-fetches: %v", errorsFromFetch)
		}

		records := fetches.Records()
		if len(records) == 0 {
			continue
		}
		record := records[0]

		err := con.processRecord(ctx, record)
		if err != nil {
			return fmt.Errorf("error from processRecord: %w", err)
		}

		err = con.client.CommitRecords(ctx, record)
		if err != nil {
			return fmt.Errorf("error from commitRecords: %w", err)
		}

	}

}

func (con *Consumer) processRecord(ctx context.Context, record *kgo.Record) error {
	var envelope event.Envelope

	err := json.Unmarshal(record.Value, &envelope)
	if err != nil {
		return fmt.Errorf("error from unmarshal: %w", err)
	}
	evt := envelope.ToEvent()

	err = con.handler.Handle(ctx, evt)
	if err != nil {
		return fmt.Errorf("error from handle: %w", err)
	}

	return nil
}
