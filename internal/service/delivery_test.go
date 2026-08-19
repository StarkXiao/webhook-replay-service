package service

import (
	"context"
	"errors"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliverRetriesAndMovesToDeadLetter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusBadGateway) }))
	defer server.Close()
	repo := repository.NewMemory()
	s := &Service{Repo: repo, Clock: clock.Real{}, MaxRetries: 1}
	e, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "test", Source: "source", TargetURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.Schedule(context.Background(), e.ID, clock.Real{}.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Deliver(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	storedTask, err := repo.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedTask.Status != domain.DeadLetterStatus {
		t.Fatalf("task status = %s", storedTask.Status)
	}
	dead, err := repo.ListDeadLetters(context.Background())
	if err != nil || len(dead) != 1 {
		t.Fatalf("dead letters = %d, err=%v", len(dead), err)
	}
}

func TestDiscardDeadLetterPersists(t *testing.T) {
	repo := repository.NewMemory()
	d := &domain.DeadLetter{ID: "dl-1", EventID: "event-1"}
	if err := repo.AddDeadLetter(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	l := Lifecycle{Repo: repo, Clock: clock.Real{}}
	if err := l.DiscardDeadLetter(context.Background(), d.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetDeadLetter(context.Background(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Discarded {
		t.Fatal("discard state was not persisted")
	}
	if !errors.Is(l.DiscardDeadLetter(context.Background(), d.ID), ErrConflict) {
		t.Fatal("expected conflict when discarding twice")
	}
}
