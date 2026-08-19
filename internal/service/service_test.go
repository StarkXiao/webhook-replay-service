package service

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReceiveRejectsInvalidJSON(t *testing.T) {
	s := &Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	if _, err := s.Receive(context.Background(), ReceiveInput{Body: []byte("bad")}); err == nil {
		t.Fatal("expected error")
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
