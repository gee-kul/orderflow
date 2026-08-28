package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
	t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092,localhost:9093")
	t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
	t.Setenv("ORDERFLOW_KAFKA_GROUP", "orderflow-order-stats")

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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ORDERFLOW_DATABASE_URL", "postgres://test")
			t.Setenv("ORDERFLOW_KAFKA_BROKERS", "localhost:9092")
			t.Setenv("ORDERFLOW_KAFKA_TOPIC", "order.events")
			t.Setenv("ORDERFLOW_KAFKA_GROUP", "orderflow_order_stats")

			t.Setenv(test.envKey, test.envValue)
			_, err := loadConfig()
			if err == nil {
				t.Fatal("must be error")
			}
		})
	}
}
