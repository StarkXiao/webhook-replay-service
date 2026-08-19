package repository

import (
	"context"
	"errors"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate event")

type Repository interface {
	CreateEvent(context.Context, *domain.WebhookEvent) error
	GetEvent(context.Context, string) (*domain.WebhookEvent, error)
	ListEvents(context.Context, domain.EventFilter) ([]domain.WebhookEvent, int, error)
	SaveTask(context.Context, *domain.DeliveryTask) error
	GetTask(context.Context, string) (*domain.DeliveryTask, error)
	ClaimTask(context.Context, string, time.Time) (*domain.DeliveryTask, error)
	ListDueTasks(context.Context, time.Time) ([]domain.DeliveryTask, error)
	UpdateEvent(context.Context, *domain.WebhookEvent) error
	UpdateTask(context.Context, *domain.DeliveryTask) error
	AddAttempt(context.Context, *domain.DeliveryAttempt) error
	Attempts(context.Context, string) ([]domain.DeliveryAttempt, error)
	AddDeadLetter(context.Context, *domain.DeadLetter) error
	GetDeadLetter(context.Context, string) (*domain.DeadLetter, error)
	ListDeadLetters(context.Context) ([]domain.DeadLetter, error)
	AddAudit(context.Context, *domain.AuditLog) error
}
type Memory struct {
	mu       sync.RWMutex
	events   map[string]*domain.WebhookEvent
	tasks    map[string]*domain.DeliveryTask
	attempts map[string][]domain.DeliveryAttempt
	dead     map[string]*domain.DeadLetter
	audits   []domain.AuditLog
}

func NewMemory() *Memory {
	return &Memory{events: map[string]*domain.WebhookEvent{}, tasks: map[string]*domain.DeliveryTask{}, attempts: map[string][]domain.DeliveryAttempt{}, dead: map[string]*domain.DeadLetter{}}
}
func (m *Memory) CreateEvent(_ context.Context, e *domain.WebhookEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.events {
		if e.IdempotencyKey != "" && x.IdempotencyKey == e.IdempotencyKey {
			return ErrDuplicate
		}
	}
	m.events[e.ID] = cloneEvent(e)
	return nil
}
func cloneEvent(e *domain.WebhookEvent) *domain.WebhookEvent {
	x := *e
	x.Headers = map[string]string{}
	for k, v := range e.Headers {
		x.Headers[k] = v
	}
	x.Body = append([]byte(nil), e.Body...)
	return &x
}
func (m *Memory) GetEvent(_ context.Context, i string) (*domain.WebhookEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.events[i]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneEvent(e), nil
}
func (m *Memory) ListEvents(_ context.Context, f domain.EventFilter) ([]domain.WebhookEvent, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.WebhookEvent{}
	for _, e := range m.events {
		if f.EventType != "" && e.EventType != f.EventType {
			continue
		}
		if f.Source != "" && e.Source != f.Source {
			continue
		}
		if f.Status != "" && string(e.Status) != f.Status {
			continue
		}
		if f.TargetURL != "" && e.TargetURL != f.TargetURL {
			continue
		}
		if f.Keyword != "" && !contains(string(e.Body), f.Keyword) && !contains(e.EventType, f.Keyword) {
			continue
		}
		out = append(out, *cloneEvent(e))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	total := len(out)
	start := f.Offset
	if start > total {
		start = total
	}
	end := total
	if f.Limit > 0 && start+f.Limit < end {
		end = start + f.Limit
	}
	return out[start:end], total, nil
}
func contains(s, q string) bool { return len(q) == 0 || len(s) >= len(q) && index(s, q) >= 0 }
func index(s, q string) int {
	for i := 0; i+len(q) <= len(s); i++ {
		if s[i:i+len(q)] == q {
			return i
		}
	}
	return -1
}
func (m *Memory) SaveTask(_ context.Context, t *domain.DeliveryTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[t.ID] = cloneTask(t)
	return nil
}
func cloneTask(t *domain.DeliveryTask) *domain.DeliveryTask { x := *t; return &x }
func (m *Memory) GetTask(_ context.Context, i string) (*domain.DeliveryTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[i]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneTask(t), nil
}
func (m *Memory) ClaimTask(_ context.Context, i string, now time.Time) (*domain.DeliveryTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[i]
	if !ok {
		return nil, ErrNotFound
	}
	if !t.Status.CanDeliver() {
		return nil, ErrDuplicate
	}
	t.Status = domain.Delivering
	t.UpdatedAt = now
	return cloneTask(t), nil
}
func (m *Memory) ListDueTasks(_ context.Context, n time.Time) ([]domain.DeliveryTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o := []domain.DeliveryTask{}
	for _, t := range m.tasks {
		if (t.Status == domain.Scheduled || t.Status == domain.Retrying || t.Status == domain.Pending) && !t.ScheduledAt.After(n) {
			o = append(o, *cloneTask(t))
		}
	}
	return o, nil
}
func (m *Memory) UpdateEvent(_ context.Context, e *domain.WebhookEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.events[e.ID]; !ok {
		return ErrNotFound
	}
	m.events[e.ID] = cloneEvent(e)
	return nil
}
func (m *Memory) UpdateTask(_ context.Context, t *domain.DeliveryTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[t.ID]; !ok {
		return ErrNotFound
	}
	m.tasks[t.ID] = cloneTask(t)
	return nil
}
func (m *Memory) AddAttempt(_ context.Context, a *domain.DeliveryAttempt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts[a.TaskID] = append(m.attempts[a.TaskID], *a)
	return nil
}
func (m *Memory) Attempts(_ context.Context, i string) ([]domain.DeliveryAttempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.DeliveryAttempt(nil), m.attempts[i]...), nil
}
func (m *Memory) AddDeadLetter(_ context.Context, d *domain.DeadLetter) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	x := *d
	x.Attempts = append([]domain.DeliveryAttempt(nil), d.Attempts...)
	m.dead[d.ID] = &x
	return nil
}
func (m *Memory) GetDeadLetter(_ context.Context, i string) (*domain.DeadLetter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.dead[i]
	if !ok {
		return nil, ErrNotFound
	}
	x := *d
	x.Attempts = append([]domain.DeliveryAttempt(nil), d.Attempts...)
	return &x, nil
}
func (m *Memory) ListDeadLetters(_ context.Context) ([]domain.DeadLetter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o := []domain.DeadLetter{}
	for _, d := range m.dead {
		o = append(o, *d)
	}
	return o, nil
}
func (m *Memory) AddAudit(_ context.Context, a *domain.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.audits = append(m.audits, *a)
	return nil
}
