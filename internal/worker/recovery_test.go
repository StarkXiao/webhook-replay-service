package worker

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"testing"
	"time"
)

func TestRecoveryReschedulesStaleDeliveringTask(t *testing.T) {
	repo := repository.NewMemory()
	event := &domain.WebhookEvent{ID: "event-1", Status: domain.Delivering, CreatedAt: time.Now()}
	if err := repo.CreateEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	task := &domain.DeliveryTask{ID: "task-1", EventID: event.ID, Status: domain.Delivering, UpdatedAt: time.Now().Add(-time.Hour), ScheduledAt: time.Now().Add(-time.Hour)}
	if err := repo.SaveTask(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if err := (Recovery{Repo: repo, Timeout: time.Minute}).Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.Retrying {
		t.Fatalf("status = %s, want retrying", got.Status)
	}
	updatedEvent, err := repo.GetEvent(context.Background(), event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedEvent.Status != domain.Retrying {
		t.Fatalf("event status = %s, want retrying", updatedEvent.Status)
	}
}
