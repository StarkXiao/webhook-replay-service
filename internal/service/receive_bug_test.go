package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
)

type createEventErrorRepository struct {
	repository.Repository
	err error
}

func (r createEventErrorRepository) CreateEvent(context.Context, *domain.WebhookEvent) error {
	return r.err
}

func TestBug004_ReceivePreservesDuplicateError(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	in := ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: "https://example.com", IdempotencyKey: "duplicate-key"}
	if _, err := s.Receive(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	_, err := s.Receive(context.Background(), in)
	if !errors.Is(err, repository.ErrDuplicate) {
		t.Fatalf("error = %v, want errors.Is(err, repository.ErrDuplicate)", err)
	}
	if !strings.Contains(err.Error(), "create event") {
		t.Fatalf("error = %v, want create-event context", err)
	}
}

func TestReceivePreservesRepositoryError(t *testing.T) {
	repositoryErr := errors.New("repository unavailable")
	s := &Service{Repo: createEventErrorRepository{err: repositoryErr}, Clock: clock.Real{}}
	_, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: "https://example.com"})
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("error = %v, want errors.Is(err, repositoryErr)", err)
	}
}
