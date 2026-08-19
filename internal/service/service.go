package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/validator"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"github.com/StarkXiao/webhook-replay-service/pkg/id"
	"net/http"
	"time"
)

type Service struct {
	Repo       repository.Repository
	Clock      clock.Clock
	MaxRetries int
}
type ReceiveInput struct {
	Headers                                      map[string]string
	Body                                         []byte
	EventType, Source, TargetURL, IdempotencyKey string
}

func (s *Service) Receive(ctx context.Context, in ReceiveInput) (*domain.WebhookEvent, error) {
	if s.Repo == nil || s.Clock == nil {
		return nil, fmt.Errorf("service is not configured")
	}
	if err := validator.All(validator.JSONBody(in.Body), validator.Required("event_type", in.EventType), validator.Required("source", in.Source), validator.URL("target_url", in.TargetURL)); err != nil {
		return nil, err
	}
	now := s.Clock.Now()
	e := &domain.WebhookEvent{ID: id.New(), Headers: in.Headers, Body: in.Body, EventType: in.EventType, Source: in.Source, TargetURL: in.TargetURL, IdempotencyKey: in.IdempotencyKey, Status: domain.Pending, MaxRetries: s.MaxRetries, ReceivedAt: now, CreatedAt: now, UpdatedAt: now}
	if err := s.Repo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}
func (s *Service) Schedule(ctx context.Context, eventID string, at time.Time) (*domain.DeliveryTask, error) {
	e, err := s.Repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !e.Status.CanDeliver() {
		return nil, ErrConflict
	}
	if err := validator.URL("target_url", e.TargetURL); err != nil {
		return nil, err
	}
	if at.IsZero() {
		at = s.Clock.Now()
	}
	t := &domain.DeliveryTask{ID: id.New(), EventID: e.ID, TargetURL: e.TargetURL, Status: domain.Scheduled, ScheduledAt: at, CreatedAt: s.Clock.Now(), UpdatedAt: s.Clock.Now()}
	e.Status = domain.Scheduled
	e.UpdatedAt = s.Clock.Now()
	if err = s.Repo.SaveTask(ctx, t); err == nil {
		err = s.Repo.UpdateEvent(ctx, e)
	}
	return t, err
}
func (s *Service) Replay(ctx context.Context, eventID, target string, at time.Time) (*domain.DeliveryTask, error) {
	e, err := s.Repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if target != "" {
		if err := validator.URL("target_url", target); err != nil {
			return nil, err
		}
		// A replay override applies to this delivery only; retain the original
		// event target so future replays preserve their recorded destination.
		if at.IsZero() {
			at = s.Clock.Now()
		}
		t := &domain.DeliveryTask{ID: id.New(), EventID: e.ID, TargetURL: target, Status: domain.Scheduled, ScheduledAt: at, CreatedAt: s.Clock.Now(), UpdatedAt: s.Clock.Now()}
		if err := s.Repo.SaveTask(ctx, t); err != nil {
			return nil, err
		}
		e.Status = domain.Scheduled
		e.UpdatedAt = s.Clock.Now()
		if err := s.Repo.UpdateEvent(ctx, e); err != nil {
			return nil, err
		}
		return t, nil
	}
	if at.IsZero() {
		at = s.Clock.Now()
	}
	return s.Schedule(ctx, e.ID, at)
}
func (s *Service) Events(ctx context.Context, f domain.EventFilter) ([]domain.WebhookEvent, int, error) {
	return s.Repo.ListEvents(ctx, f)
}
func (s *Service) Deliver(ctx context.Context, t *domain.DeliveryTask) error {
	if t == nil {
		return fmt.Errorf("delivery task is required")
	}
	now := s.Clock.Now()
	claimed, err := s.Repo.ClaimTask(ctx, t.ID, now)
	if err != nil {
		return err
	}
	t = claimed
	e, err := s.Repo.GetEvent(ctx, t.EventID)
	if err != nil {
		return err
	}
	target := t.TargetURL
	if target == "" {
		target = e.TargetURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(e.Body))
	if err != nil {
		return err
	}
	for k, v := range e.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	a := &domain.DeliveryAttempt{ID: id.New(), TaskID: t.ID, Attempt: t.Attempt + 1, StartedAt: s.Clock.Now(), FinishedAt: s.Clock.Now()}
	if err != nil {
		a.Error = err.Error()
		t.Error = err.Error()
		t.Status = domain.Failed
	} else {
		a.StatusCode = resp.StatusCode
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Status = domain.Success
		} else {
			t.Status = domain.Failed
			t.Error = fmt.Sprintf("status %d", resp.StatusCode)
		}
	}
	if err := s.Repo.AddAttempt(ctx, a); err != nil {
		return err
	}
	t.Attempt++
	t.UpdatedAt = s.Clock.Now()
	if err := s.Repo.UpdateTask(ctx, t); err != nil {
		return err
	}
	e.Status = t.Status
	e.LastError = t.Error
	e.UpdatedAt = t.UpdatedAt
	return s.Repo.UpdateEvent(ctx, e)
}
