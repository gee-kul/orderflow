package event

import (
	"encoding/json"
	"time"
)

type Envelope struct {
	EventID       string          `json:"event_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

func NewEnvelope(evt Event) Envelope{
	envelope := Envelope{EventID: evt.ID, AggregateType: evt.AggregateType,
	AggregateID: evt.AggregateID, EventType: evt.EventType, Payload: evt.Payload,
	CreatedAt: evt.CreatedAt}
	return envelope
}

func (env Envelope) ToEvent() Event{
	evt := Event{ID: env.EventID, AggregateType: env.AggregateType,
	AggregateID: env.AggregateID, EventType: env.EventType, Payload: env.Payload,
	CreatedAt: env.CreatedAt, PublishedAt: nil}
	return evt
}
