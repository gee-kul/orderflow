package outbox

import (
	"context"
	"fmt"
)

type Worker struct {
	repository Repository
	publisher  Publisher
	batchSize  int
}

func NewWorker(repo Repository, pub Publisher, size int) *Worker {
	var worker Worker
	worker.repository = repo
	worker.publisher = pub
	worker.batchSize = size
	return &worker
}

func (w *Worker) RunOnce(ctx context.Context) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	events, err := w.repository.FetchUnpublished(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("не удалось получить неопубликованные события: %w", err)
	}

	for _, event := range events {
		err := ctx.Err()
		if err != nil {
			return err
		}

		err = w.publisher.Publish(ctx, event)
		if err != nil {
			return fmt.Errorf("не удалось опубликовать событие: %v, %w", event.ID, err)
		}

		err = w.repository.MarkPublished(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("не удалось отметить событие: %v, %w", event.ID, err)
		}
	}
	return nil
}
