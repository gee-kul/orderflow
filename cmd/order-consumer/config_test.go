package main

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
	t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092,localhost:9093")
	t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
	t.Setenv("ORDERFLOW_KAFKA_GROUP", "orderflow-order-stats")
	t.Setenv("ORDERFLOW_CONSUMER_MAX_ATTEMPTS", "")
	t.Setenv("ORDERFLOW_CONSUMER_INITIAL_BACKOFF", "")
	t.Setenv("ORDERFLOW_CONSUMER_MAX_BACKOFF", "")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("error from loadConfig %v", err)
	}

	if cfg.databaseURL != "postgres://test" {
		t.Errorf("databaseURL must be postgres://test, but received %v", cfg.databaseURL)
	}

	if len(cfg.kafkaBrokers) != 2 {
		t.Fatalf("len kafkaBrokers must be 2, but received %v", len(cfg.kafkaBrokers))
	}
	if cfg.kafkaBrokers[0] != "localhost:9092" {
		t.Errorf("kafkaBrokers must be localhost:9092, but received %v", cfg.kafkaBrokers[0])
	}
	if cfg.kafkaBrokers[1] != "localhost:9093" {
		t.Errorf("kafkaBrokers must be localhost:9093, but received %v", cfg.kafkaBrokers[1])
	}

	if cfg.kafkaTopic != "order.events" {
		t.Errorf("kafkaTopic must be order.events, but received %v", cfg.kafkaTopic)
	}

	if cfg.kafkaGroup != "orderflow-order-stats" {
		t.Errorf("kafkaGroup must be orderflow-order-stats, but received %v", cfg.kafkaGroup)
	}

	if cfg.maxAttempts != 3 {
		t.Errorf("maxAttempts must be 3, but received: %v", cfg.maxAttempts)
	}

	if cfg.initialBackoff != 500*time.Millisecond {
		t.Errorf("initialBackoff must be 500ms, but received: %v", cfg.initialBackoff)
	}

	if cfg.maxBackoff != 5*time.Second {
		t.Errorf("maxBackoff must be 5s, but received: %v", cfg.maxBackoff)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
	}{
		{
			name:     "empty databaseURL",
			envKey:   "ORDERFLOW_DATABASE_URL",
			envValue: " ",
		},
		{
			name:     "empty brokers",
			envKey:   "ORDERFLOW_KAFKA_BROKERS",
			envValue: " ",
		},
		{
			name:     "empty topic",
			envKey:   "ORDERFLOW_KAFKA_TOPIC",
			envValue: " ",
		},
		{
			name:     "empty group",
			envKey:   "ORDERFLOW_KAFKA_GROUP",
			envValue: " ",
		},
		{
			name:     "not number",
			envKey:   "ORDERFLOW_CONSUMER_MAX_ATTEMPTS",
			envValue: "abc",
		},
		{
			name:     "zero attempts",
			envKey:   "ORDERFLOW_CONSUMER_MAX_ATTEMPTS",
			envValue: "0",
		},
		{
			name:     "-1 attempts",
			envKey:   "ORDERFLOW_CONSUMER_MAX_ATTEMPTS",
			envValue: "-1",
		},
		{
			name:     "not backoff",
			envKey:   "ORDERFLOW_CONSUMER_INITIAL_BACKOFF",
			envValue: "abc",
		},
		{
			name:     "-1 backoff",
			envKey:   "ORDERFLOW_CONSUMER_INITIAL_BACKOFF",
			envValue: "-1s",
		},
		{
			name:     "not max backoff",
			envKey:   "ORDERFLOW_CONSUMER_MAX_BACKOFF",
			envValue: "abc",
		},
		{
			name:     "-1 max backoff",
			envKey:   "ORDERFLOW_CONSUMER_MAX_BACKOFF",
			envValue: "-1s",
		},
		{
			name:     "max < initial",
			envKey:   "ORDERFLOW_CONSUMER_MAX_BACKOFF",
			envValue: "100ms",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
			t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092")
			t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
			t.Setenv("ORDERFLOW_KAFKA_GROUP", "orderflow_order_stats")
			t.Setenv("ORDERFLOW_CONSUMER_MAX_ATTEMPTS", "3")
			t.Setenv("ORDERFLOW_CONSUMER_INITIAL_BACKOFF", "500ms")
			t.Setenv("ORDERFLOW_CONSUMER_MAX_BACKOFF", "5s")

			t.Setenv(test.envKey, test.envValue)
			_, err := loadConfig()
			if err == nil {
				t.Fatal("must be error")
			}
		})
	}
}

func TestLoadConfigRetryOverrides(t *testing.T) {
	t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
	t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092,localhost:9093")
	t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
	t.Setenv("ORDERFLOW_KAFKA_GROUP", "orderflow-order-stats")
	t.Setenv("ORDERFLOW_CONSUMER_MAX_ATTEMPTS", "5")
	t.Setenv("ORDERFLOW_CONSUMER_INITIAL_BACKOFF", "200ms")
	t.Setenv("ORDERFLOW_CONSUMER_MAX_BACKOFF", "2s")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("error from loadConfig %v", err)
	}

	if cfg.databaseURL != "postgres://test" {
		t.Errorf("databaseURL must be postgres://test, but received %v", cfg.databaseURL)
	}

	if cfg.maxAttempts != 5 {
		t.Errorf("maxAttempts must be 5, but received: %v", cfg.maxAttempts)
	}

	if cfg.initialBackoff != 200*time.Millisecond {
		t.Errorf("initialBackoff must be 200ms, but received: %v", cfg.initialBackoff)
	}

	if cfg.maxBackoff != 2*time.Second {
		t.Errorf("maxBackoff must be 2s, but received: %v", cfg.maxBackoff)
	}
}
