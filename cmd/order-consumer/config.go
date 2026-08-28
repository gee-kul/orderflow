package main

import (
	"fmt"
	"os"
	"strings"
)

type config struct {
	databaseURL  string
	kafkaBrokers []string
	kafkaTopic   string
	kafkaGroup   string
}

func loadConfig() (config, error) {
	var cfg config

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

	return cfg, nil

}
