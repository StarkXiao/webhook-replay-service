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
	if s.Every <= 0 {
		s.Every = time.Second
	}
	t := time.NewTicker(s.Every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			tasks, err := s.Repo.ListDueTasks(ctx, now)
			if err != nil {
				continue
			}
			for _, task := range tasks {
				if !s.Pool.Submit(ctx, task) {
					return
				}
			}
		}
	}
}
