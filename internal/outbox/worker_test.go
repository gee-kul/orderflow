package outbox

import (
	"context"
	"errors"
	"testing"

	"github.com/gee-kul/orderflow/internal/event"
)

type fakeRepository struct {
	events    []event.Event
	fetchErr  error
	markErr   error
	markedIDs []string
}

type fakePublisher struct {
	published  []event.Event
	publishErr error
}

var (
	ErrPublish = errors.New("error to publish")
	ErrMark    = errors.New("error to mark")
)

func (repo *fakeRepository) FetchUnpublished(ctx context.Context, limit int) ([]event.Event, error) {
	return repo.events, repo.fetchErr
}

func (repo *fakeRepository) MarkPublished(ctx context.Context, id string) error {
	repo.markedIDs = append(repo.markedIDs, id)
	return repo.markErr
}

func (pub *fakePublisher) Publish(ctx context.Context, evt event.Event) error {
	pub.published = append(pub.published, evt)
	return pub.publishErr
}

func TestNoEvents(t *testing.T) {
	publisher := fakePublisher{}
	repo := fakeRepository{}
	worker := NewWorker(&repo, &publisher, 1)

	err := worker.RunOnce(t.Context())
	if err != nil {
		t.Errorf("не должно быть ошибок, получили %v", err)
	}

	if len(publisher.published) != 0 {
		t.Fatalf("published должен быть пустым, получили длину %v", len(publisher.published))
	}

	if len(repo.markedIDs) != 0 {
		t.Fatalf("markedIDs должен быть пустым, получили длину %v", len(repo.markedIDs))
	}
}

func TestSuccess(t *testing.T) {
	evt := event.Event{ID: "id-1"}
	publisher := fakePublisher{}
	repo := fakeRepository{}

	repo.events = append(repo.events, evt)
	worker := NewWorker(&repo, &publisher, 1)

	err := worker.RunOnce(t.Context())
	if err != nil {
		t.Fatalf("не должно быть ошибок, получили %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatalf("published должен содержать одно событие, получили длину %v", len(publisher.published))
	}

	if len(repo.markedIDs) != 1 {
		t.Fatalf("markedIDs должен содержать один айди, получили длину %v", len(repo.markedIDs))
	}

	if publisher.published[0].ID != evt.ID {
		t.Fatalf("айди publisher не совпадают, должно быть %v", evt.ID)
	}

	if repo.markedIDs[0] != evt.ID {
		t.Fatalf("айди repo не совпадают, должно быть %v", evt.ID)
	}
}

func TestErrPublish(t *testing.T) {
	evt := event.Event{ID: "id-1"}
	publisher := fakePublisher{publishErr: ErrPublish}
	repo := fakeRepository{}

	repo.events = append(repo.events, evt)
	worker := NewWorker(&repo, &publisher, 1)

	err := worker.RunOnce(t.Context())
	if !errors.Is(err, ErrPublish) {
		t.Errorf("должна быть ошибка, получили %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatal("published должен был получить событие")
	}

	if len(repo.markedIDs) != 0 {
		t.Fatalf("markedIDs должен быть пустым, получили длину %v", len(repo.markedIDs))
	}
}

func TestErrMark(t *testing.T) {
	evt := event.Event{ID: "id-1"}
	publisher := fakePublisher{}
	repo := fakeRepository{markErr: ErrMark}

	repo.events = append(repo.events, evt)
	worker := NewWorker(&repo, &publisher, 1)

	err := worker.RunOnce(t.Context())
	if !errors.Is(err, ErrMark) {
		t.Errorf("должна быть ошибка, получили %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatal("published должен был получить событие")
	}

	if len(repo.markedIDs) != 1 {
		t.Fatalf("markedIDs должен содержать один айди, получили длину %v", len(repo.markedIDs))
	}
}
