package worker

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"time"
)

type Recovery struct {
	Repo    repository.Repository
	Timeout time.Duration
}

func (r Recovery) Recover(ctx context.Context) error {
	tasks, err := r.Repo.ListDueTasks(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Status == domain.Delivering && time.Since(task.UpdatedAt) > r.Timeout {
			task.Status = domain.Retrying
			task.ScheduledAt = time.Now().UTC()
			task.Error = "recovered after delivery timeout"
			if err := r.Repo.UpdateTask(ctx, &task); err != nil {
				return err
			}
		}
	}
	return nil
}

type Maintenance struct {
	Repo         repository.Repository
	ArchiveAfter time.Duration
}

func (m Maintenance) Run(ctx context.Context) error {
	_, err := m.Repo.ListDueTasks(ctx, time.Now().UTC())
	return err
}
func WorkerName(index int) string {
	return "delivery-worker-" + time.Now().UTC().Format("20060102") + "-" + string(rune('0'+index%10))
}
