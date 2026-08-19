package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"sync"
	"time"
)

var ErrConflict = errors.New("resource conflict")

type Lifecycle struct {
	Repo  repository.Repository
	Clock interface{ Now() time.Time }
}

func (l Lifecycle) MarkDelivering(ctx context.Context, taskID string) (*domain.DeliveryTask, error) {
	t, err := l.Repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !t.Status.CanDeliver() {
		return nil, ErrConflict
	}
	t.Status = domain.Delivering
	t.UpdatedAt = l.Clock.Now()
	if err := l.Repo.UpdateTask(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
func (l Lifecycle) Complete(ctx context.Context, taskID string, status int, summary string) error {
	t, err := l.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	e, err := l.Repo.GetEvent(ctx, t.EventID)
	if err != nil {
		return err
	}
	now := l.Clock.Now()
	a := &domain.DeliveryAttempt{ID: fmt.Sprintf("attempt-%s-%d", t.ID, t.Attempt+1), TaskID: t.ID, Attempt: t.Attempt + 1, StatusCode: status, ResponseSummary: summary, StartedAt: now, FinishedAt: now}
	if status >= 200 && status < 300 {
		t.Status = domain.Success
		e.Status = domain.Success
		e.LastError = ""
	} else {
		t.Status = domain.Failed
		e.Status = domain.Failed
		t.Error = fmt.Sprintf("delivery returned %d", status)
		e.LastError = t.Error
	}
	t.Attempt++
	t.UpdatedAt = now
	e.UpdatedAt = now
	if err = l.Repo.AddAttempt(ctx, a); err != nil {
		return err
	}
	if err = l.Repo.UpdateTask(ctx, t); err != nil {
		return err
	}
	return l.Repo.UpdateEvent(ctx, e)
}
func (l Lifecycle) Fail(ctx context.Context, taskID string, cause error, next time.Time, maxRetries int) error {
	t, err := l.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	e, err := l.Repo.GetEvent(ctx, t.EventID)
	if err != nil {
		return err
	}
	now := l.Clock.Now()
	t.Attempt++
	t.Error = cause.Error()
	t.UpdatedAt = now
	e.RetryCount++
	e.LastError = cause.Error()
	e.UpdatedAt = now
	if t.Attempt >= maxRetries {
		t.Status = domain.DeadLetterStatus
		e.Status = domain.DeadLetterStatus
		d := &domain.DeadLetter{ID: "dlq-" + t.ID, EventID: e.ID, Reason: cause.Error(), CreatedAt: now, UpdatedAt: now}
		if err = l.Repo.AddDeadLetter(ctx, d); err != nil {
			return err
		}
	} else {
		t.Status = domain.Retrying
		t.ScheduledAt = next
		e.Status = domain.Retrying
	}
	if err = l.Repo.UpdateTask(ctx, t); err != nil {
		return err
	}
	return l.Repo.UpdateEvent(ctx, e)
}
func (l Lifecycle) Reschedule(ctx context.Context, taskID string, at time.Time) error {
	t, err := l.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if t.Status == domain.Success || t.Status == domain.Cancelled {
		return ErrConflict
	}
	t.ScheduledAt = at
	t.Status = domain.Scheduled
	t.UpdatedAt = l.Clock.Now()
	return l.Repo.UpdateTask(ctx, t)
}
func (l Lifecycle) DiscardDeadLetter(ctx context.Context, id string) error {
	d, err := l.Repo.GetDeadLetter(ctx, id)
	if err != nil {
		return err
	}
	if d.Discarded {
		return ErrConflict
	}
	d.Discard()
	d.UpdatedAt = l.Clock.Now()
	return l.Repo.AddAudit(ctx, &domain.AuditLog{ID: "audit-" + id, Action: "dead_letter.discarded", EntityType: "dead_letter", EntityID: id, CreatedAt: l.Clock.Now()})
}

type Gate struct {
	mu     sync.Mutex
	active map[string]time.Time
}

func NewGate() *Gate { return &Gate{active: map[string]time.Time{}} }
func (g *Gate) Acquire(key string, ttl time.Duration) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if until, ok := g.active[key]; ok && until.After(time.Now()) {
		return false
	}
	g.active[key] = time.Now().Add(ttl)
	return true
}
func (g *Gate) Release(key string) { g.mu.Lock(); defer g.mu.Unlock(); delete(g.active, key) }
func (g *Gate) Cleanup(now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for k, v := range g.active {
		if !v.After(now) {
			delete(g.active, k)
		}
	}
}
