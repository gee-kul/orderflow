package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	databaseURL  string
	kafkaBrokers []string
	kafkaTopic   string
	batchSize    int
	pollInterval time.Duration
}

func loadConfig() (config, error) {
	var cfg config
	cfg.batchSize = 100
	cfg.pollInterval = time.Second

	database, ok := os.LookupEnv("ORDERFLOW_DATABASE_URL")
	if !ok || strings.TrimSpace(database) == "" {
		return config{}, fmt.Errorf("чтение ORDERFLOW_DATABASE_URL не успешно")
	}
	cfg.databaseURL = strings.TrimSpace(database)

	brokers, ok := os.LookupEnv("ORDERFLOW_KAFKA_BROKERS")
	if !ok || strings.TrimSpace(brokers) == "" {
		return config{}, fmt.Errorf("чтение ORDERFLOW_KAFKA_BROKERS не успешно")
	}

	brokersSplit := strings.Split(brokers, ",")
	cfg.kafkaBrokers = brokersSplit

	topic, ok := os.LookupEnv("ORDERFLOW_KAFKA_TOPIC")
	if !ok || strings.TrimSpace(topic) == "" {
		return config{}, fmt.Errorf("чтение ORDERFLOW_KAFKA_TOPIC не успешно")
	}
	cfg.kafkaTopic = strings.TrimSpace(topic)

	batch, ok := os.LookupEnv("ORDERFLOW_OUTBOX_BATCH_SIZE")
	if ok && strings.TrimSpace(batch) != "" {
		batchInt, err := strconv.Atoi(strings.TrimSpace(batch))
		if err != nil {
			return config{}, fmt.Errorf("не удалось преобразовать batch %w", err)
		}
		if batchInt <= 0 {
			return config{}, fmt.Errorf("batchSize меньше или равен нуля")
		}
		cfg.batchSize = batchInt
	}

	poll, ok := os.LookupEnv("ORDERFLOW_OUTBOX_POLL_INTERVAL")
	if ok && strings.TrimSpace(poll) != "" {
		pollInterval, err := time.ParseDuration(strings.TrimSpace(poll))
		if err != nil {
			return config{}, fmt.Errorf("не удалось преобразовать pollInterval %w", err)
		}
		if pollInterval <= 0 {
			return config{}, fmt.Errorf("pollInterval меньше или равен нуля")
		}
		cfg.pollInterval = pollInterval
	}

	return cfg, nil
}
