package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	databaseURL    string
	kafkaBrokers   []string
	kafkaTopic     string
	kafkaGroup     string
	maxAttempts    int
	initialBackoff time.Duration
	maxBackoff     time.Duration
}

func loadConfig() (config, error) {
	var cfg config

	cfg.maxAttempts = 3
	cfg.initialBackoff = 500 * time.Millisecond
	cfg.maxBackoff = 5 * time.Second

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

	group, ok := os.LookupEnv("ORDERFLOW_KAFKA_GROUP")
	if !ok || strings.TrimSpace(group) == "" {
		return config{}, fmt.Errorf("чтение ORDERFLOW_KAFKA_GROUP не успешно")
	}
	cfg.kafkaGroup = strings.TrimSpace(group)

	maxAttempts := os.Getenv("ORDERFLOW_CONSUMER_MAX_ATTEMPTS")
	maxAttempts = strings.TrimSpace(maxAttempts)
	if maxAttempts != "" {
		maxAttemptsInt, err := strconv.Atoi(maxAttempts)
		if err != nil {
			return config{}, fmt.Errorf("error from atoi: %w", err)
		}
		cfg.maxAttempts = maxAttemptsInt
	}

	initialBackoff := os.Getenv("ORDERFLOW_CONSUMER_INITIAL_BACKOFF")
	initialBackoff = strings.TrimSpace(initialBackoff)
	if initialBackoff != "" {
		initialBackoffTime, err := time.ParseDuration(initialBackoff)
		if err != nil {
			return config{}, fmt.Errorf("error from parseDuration initial: %w", err)
		}
		cfg.initialBackoff = initialBackoffTime
	}

	maxBackoff := os.Getenv("ORDERFLOW_CONSUMER_MAX_BACKOFF")
	maxBackoff = strings.TrimSpace(maxBackoff)
	if maxBackoff != "" {
		maxBackoffTime, err := time.ParseDuration(maxBackoff)
		if err != nil {
			return config{}, fmt.Errorf("error from parseDuration max: %w", err)
		}
		cfg.maxBackoff = maxBackoffTime
	}

	if cfg.maxAttempts < 1 {
		return config{}, fmt.Errorf("error maxAttempts < 1")
	}
	if cfg.initialBackoff < 0 {
		return config{}, fmt.Errorf("error initialBackoff < 0")
	}
	if cfg.maxBackoff < 0 {
		return config{}, fmt.Errorf("error maxBackoff < 0")
	}
	if cfg.maxBackoff < cfg.initialBackoff {
		return config{}, fmt.Errorf("error maxBackoff < initialBackoff")
	}

	return cfg, nil

}
