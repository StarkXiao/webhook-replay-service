package service

import (
	"context"
	"errors"
	"testing"

	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
)

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
}
