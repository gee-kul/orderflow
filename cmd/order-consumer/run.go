package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gee-kul/orderflow/internal/consumer/kafka"
	"github.com/gee-kul/orderflow/internal/orderstats/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error from loadConfig: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	pool, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("error from new pool: %w", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("error from ping pool: %w", err)
	}

	handler := postgres.NewHandler(pool)

	consumer, err := kafka.NewConsumer(cfg.kafkaBrokers, cfg.kafkaTopic,
		cfg.kafkaGroup, handler)
	if err != nil {
		return fmt.Errorf("error from newConsumer: %w", err)
	}
	defer consumer.Close()

	err = consumer.Run(ctx)
	if err != nil {
		return fmt.Errorf("error from run: %w", err)
	}

	return nil
}
