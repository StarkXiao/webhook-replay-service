package service

import (
	"context"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/pkg/id"
	"sync"
	"time"
)

type ReplayRequest struct {
	EventIDs       []string   `json:"event_ids"`
	TargetURL      string     `json:"target_url"`
	At             *time.Time `json:"at"`
	MaxConcurrency int        `json:"max_concurrency"`
}
type ReplayResult struct {
	BatchID  string   `json:"batch_id"`
	TaskIDs  []string `json:"task_ids"`
	Accepted int      `json:"accepted"`
	Failed   int      `json:"failed"`
}

func (s *Service) ReplayBatch(ctx context.Context, in ReplayRequest) (ReplayResult, error) {
	if len(in.EventIDs) == 0 {
		return ReplayResult{}, fmt.Errorf("event_ids cannot be empty")
	}
	if in.MaxConcurrency <= 0 {
		in.MaxConcurrency = 4
	}
	out := ReplayResult{BatchID: id.New()}
	sem := make(chan struct{}, in.MaxConcurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, eid := range in.EventIDs {
		eid := eid
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			at := time.Time{}
			if in.At != nil {
				at = *in.At
			}
			t, err := s.Replay(ctx, eid, in.TargetURL, at)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				out.Failed++
			} else {
				out.Accepted++
				out.TaskIDs = append(out.TaskIDs, t.ID)
			}
		}()
	}
	wg.Wait()
	return out, nil
}
func (s *Service) CancelTask(ctx context.Context, taskID string) error {
	t, err := s.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if err = t.Transition(domain.Cancelled); err != nil {
		return err
	}
	t.UpdatedAt = s.Clock.Now()
	return s.Repo.UpdateTask(ctx, t)
}
func (s *Service) RetryTask(ctx context.Context, taskID string) error {
	t, err := s.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	t.Status = domain.Retrying
	t.ScheduledAt = s.Clock.Now()
	t.UpdatedAt = s.Clock.Now()
	return s.Repo.UpdateTask(ctx, t)
}
