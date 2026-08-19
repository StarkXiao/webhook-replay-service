package scheduler

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/worker"
	"time"
)

type Scheduler struct {
	Repo  repository.Repository
	Pool  *worker.Pool
	Every time.Duration
}

func (s *Scheduler) Run(ctx context.Context) {
	t := time.NewTicker(s.Every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			tasks, _ := s.Repo.ListDueTasks(ctx, now)
			for _, task := range tasks {
				s.Pool.Submit(task)
			}
		}
	}
}
