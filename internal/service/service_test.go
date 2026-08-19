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
	"time"
)

func TestReceiveRejectsInvalidJSON(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	if _, err := s.Receive(context.Background(), ReceiveInput{Body: []byte("bad")}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelTaskUpdatesEventStatus(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	e, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.Schedule(context.Background(), e.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CancelTask(context.Background(), task.ID); err != nil {
		t.Fatal(err)
	}
	storedTask, err := s.Repo.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedEvent, err := s.Repo.GetEvent(context.Background(), e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedTask.Status != domain.Cancelled || storedEvent.Status != domain.Cancelled {
		t.Fatalf("task=%s event=%s, want both cancelled", storedTask.Status, storedEvent.Status)
	}
}

func TestRetryTaskUpdatesEventAndRejectsNonFailedTask(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	e, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.Schedule(context.Background(), e.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RetryTask(context.Background(), task.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("RetryTask() error = %v, want ErrConflict", err)
	}
	task.Status = domain.Failed
	if err := s.Repo.UpdateTask(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if err := s.RetryTask(context.Background(), task.ID); err != nil {
		t.Fatal(err)
	}
	storedTask, err := s.Repo.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedEvent, err := s.Repo.GetEvent(context.Background(), e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedTask.Status != domain.Retrying || storedEvent.Status != domain.Retrying {
		t.Fatalf("task=%s event=%s, want both retrying", storedTask.Status, storedEvent.Status)
	}
}

func TestReplayBatchReturnsCancelledContext(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := s.ReplayBatch(ctx, ReplayRequest{EventIDs: []string{"missing"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReplayBatch() error = %v, want context.Canceled", err)
	}
	if result.Accepted != 0 {
		t.Fatalf("accepted = %d, want 0", result.Accepted)
	}
}

func TestReceiveValidatesRequiredMetadata(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	if _, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{}`)}); err == nil {
		t.Fatal("expected metadata validation error")
	}
}

func TestReplayUsesTargetOverrideForDelivery(t *testing.T) {
	var originalCalls, replayCalls int
	original := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { originalCalls++ }))
	defer original.Close()
	replay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replayCalls++
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer replay.Close()

	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}, MaxRetries: 3}
	e, err := s.Receive(context.Background(), ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: original.URL})
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.Replay(context.Background(), e.ID, replay.URL, clock.Real{}.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Deliver(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if replayCalls != 1 || originalCalls != 0 {
		t.Fatalf("calls original=%d replay=%d, want 0 and 1", originalCalls, replayCalls)
	}
	stored, err := s.Repo.GetEvent(context.Background(), e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.TargetURL != original.URL || stored.Status != "success" {
		t.Fatalf("event = %#v", stored)
	}
}
