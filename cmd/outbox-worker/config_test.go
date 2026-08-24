package main

import (
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
	t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092")
	t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
	t.Setenv("ORDERFLOW_OUTBOX_BATCH_SIZE", "")
	t.Setenv("ORDERFLOW_OUTBOX_POLL_INTERVAL", "")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("error from loadConfig %v", err)
	}
	if cfg.batchSize != 100 {
		t.Errorf("batchSize must be 100, but replay %v", cfg.batchSize)
	}

	if cfg.pollInterval != time.Second {
		t.Errorf("pollInterval must be second, but replay %v", cfg.pollInterval)
	}

	if cfg.databaseURL != "postgres://test" {
		t.Errorf("databaseURL must be postgres://test, but replay %v", cfg.databaseURL)
	}

	if len(cfg.kafkaBrokers) != 1 {
		t.Fatalf("len kafkaBrokers must be 1, but replay %v", len(cfg.kafkaBrokers))
	}

	if cfg.kafkaBrokers[0] != "localhost:9092" {
		t.Errorf("kafkaBrokers must be localhost:9092, but replay %v", cfg.kafkaBrokers[0])
	}

	if cfg.kafkaTopic != "order.events" {
		t.Errorf("kafkaTopic must be order.events, but replay %v", cfg.kafkaTopic)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
	t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092")
	t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
	t.Setenv("ORDERFLOW_OUTBOX_BATCH_SIZE", "50")
	t.Setenv("ORDERFLOW_OUTBOX_POLL_INTERVAL", "2s")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("error from loadConfig %v", err)
	}
	if cfg.batchSize != 50 {
		t.Errorf("batchSize must be 50, but replay %v", cfg.batchSize)
	}

	if cfg.pollInterval != time.Second*2 {
		t.Errorf("pollInterval must be 2 seconds, but replay %v", cfg.pollInterval)
	}

	if cfg.databaseURL != "postgres://test" {
		t.Errorf("databaseURL must be postgres://test, but replay %v", cfg.databaseURL)
	}

	if len(cfg.kafkaBrokers) != 1 {
		t.Fatalf("len kafkaBrokers must be 1, but replay %v", len(cfg.kafkaBrokers))
	}
	if cfg.kafkaBrokers[0] != "localhost:9092" {
		t.Errorf("kafkaBrokers must be localhost:9092, but replay %v", cfg.kafkaBrokers[0])
	}

	if cfg.kafkaTopic != "order.events" {
		t.Errorf("kafkaTopic must be order.events, but replay %v", cfg.kafkaTopic)
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
			name:     "invalid batch",
			envKey:   "ORDERFLOW_OUTBOX_BATCH_SIZE",
			envValue: "abc",
		},
		{
			name:     "zero batch",
			envKey:   "ORDERFLOW_OUTBOX_BATCH_SIZE",
			envValue: "0",
		},
		{
			name:     "invalid interval",
			envKey:   "ORDERFLOW_OUTBOX_POLL_INTERVAL",
			envValue: "abc",
		},
		{
			name:     "zero interval",
			envKey:   "ORDERFLOW_OUTBOX_POLL_INTERVAL",
			envValue: "0s",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
			t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092")
			t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
			t.Setenv("ORDERFLOW_OUTBOX_BATCH_SIZE", "50")
			t.Setenv("ORDERFLOW_OUTBOX_POLL_INTERVAL", "2s")

			t.Setenv(test.envKey, test.envValue)
			_, err := loadConfig()
			if err == nil {
				t.Fatal("must be error")
			}
		})
	}
}
