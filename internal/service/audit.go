package service

import (
	"context"
	"encoding/json"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/pkg/id"
	"time"
)

type AuditService struct {
	Repo interface {
		AddAudit(context.Context, *domain.AuditLog) error
	}
}

func (a AuditService) Record(ctx context.Context, action, kind, entity string, detail any) error {
	b, _ := json.Marshal(detail)
	return a.Repo.AddAudit(ctx, &domain.AuditLog{ID: id.New(), Action: action, EntityType: kind, EntityID: entity, Detail: string(b), CreatedAt: time.Now().UTC()})
}

type AuditEntry struct {
	Action     string
	EntityType string
	EntityID   string
	Detail     string
	At         time.Time
}

func AuditActions() []string {
	return []string{"event.received", "event.scheduled", "event.replayed", "task.delivered", "task.retried", "task.cancelled", "dead_letter.retried", "dead_letter.discarded"}
}
