package scheduler

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/worker"
	"testing"
	"time"
)

func TestRunStopsWhileSubmitIsBlocked(t *testing.T) {
	repo := repository.NewMemory()
	if err := repo.SaveTask(context.Background(), &domain.DeliveryTask{ID: "task-1", Status: domain.Scheduled, ScheduledAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	pool := &worker.Pool{Jobs: make(chan domain.DeliveryTask, 1)}
	pool.Jobs <- domain.DeliveryTask{ID: "occupied"}
	runner := &Scheduler{Repo: repo, Pool: pool, Every: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runner.Run(ctx) }()
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancellation")
	}
}
