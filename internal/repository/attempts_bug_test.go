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

func TestBug003_AttemptsResultCanBeReorderedWithoutChangingStorage(t *testing.T) {
	repo := NewMemory()
	for _, attempt := range []domain.DeliveryAttempt{
		{ID: "attempt-1", TaskID: "task-1", Attempt: 1, StatusCode: 500},
		{ID: "attempt-2", TaskID: "task-1", Attempt: 2, StatusCode: 200},
	} {
		if err := repo.AddAttempt(context.Background(), &attempt); err != nil {
			t.Fatal(err)
		}
	}

	attempts, err := repo.Attempts(context.Background(), "task-1")
	if err != nil {
		t.Fatal(err)
	}
	attempts[0], attempts[1] = attempts[1], attempts[0]
	attempts = append(attempts, domain.DeliveryAttempt{ID: "caller-only"})

	stored, err := repo.Attempts(context.Background(), "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 2 || stored[0].ID != "attempt-1" || stored[1].ID != "attempt-2" {
		t.Fatalf("stored attempts changed after caller mutation: %#v", stored)
	}
}
