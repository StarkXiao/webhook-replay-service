package deadletter

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/id"
	"time"
)

func Move(ctx context.Context, r repository.Repository, eventID, reason string) (*domain.DeadLetter, error) {
	a, _ := r.Attempts(ctx, eventID)
	d := &domain.DeadLetter{ID: id.New(), EventID: eventID, Reason: reason, Attempts: a, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return d, r.AddDeadLetter(ctx, d)
}
