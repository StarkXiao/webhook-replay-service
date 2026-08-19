package repository

import (
	"context"
	"testing"

	"github.com/StarkXiao/webhook-replay-service/internal/domain"
)

func TestBug003_AttemptsReturnsIndependentSlice(t *testing.T) {
	repo := NewMemory()
	if err := repo.AddAttempt(context.Background(), &domain.DeliveryAttempt{ID: "attempt-1", TaskID: "task-1", StatusCode: 500}); err != nil {
		t.Fatal(err)
	}
	attempts, err := repo.Attempts(context.Background(), "task-1")
	if err != nil {
		t.Fatal(err)
	}
	attempts[0].StatusCode = 204
	stored, err := repo.Attempts(context.Background(), "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored[0].StatusCode != 500 {
		t.Fatalf("stored status = %d, want 500", stored[0].StatusCode)
	}
}
