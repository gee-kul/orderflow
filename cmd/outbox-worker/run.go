package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gee-kul/orderflow/internal/outbox"
	"github.com/gee-kul/orderflow/internal/outbox/kafka"
	"github.com/gee-kul/orderflow/internal/outbox/postgres"
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

	publisher, err := kafka.NewPublisher(cfg.kafkaBrokers, cfg.kafkaTopic)
	if err != nil {
		return fmt.Errorf("error from NewPublisher: %w", err)
	}
	defer publisher.Close()

	repo := postgres.NewRepository(pool)
	worker := outbox.NewWorker(repo, publisher, cfg.batchSize)

	tick := time.NewTicker(cfg.pollInterval)
	defer tick.Stop()

	for {
		err := worker.RunOnce(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("error from RunOnce: %v", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			continue
		}
	}
}
