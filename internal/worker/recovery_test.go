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
	task := &domain.DeliveryTask{ID: "task-1", Status: domain.Delivering, UpdatedAt: time.Now().Add(-time.Hour), ScheduledAt: time.Now().Add(-time.Hour)}
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
}
