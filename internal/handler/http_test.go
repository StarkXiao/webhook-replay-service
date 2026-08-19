package handler

import (
	"bytes"
	"context"
	"errors"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingCreateEventRepository struct{ repository.Repository }

func (failingCreateEventRepository) CreateEvent(context.Context, *domain.WebhookEvent) error {
	return errors.New("repository unavailable")
}

func TestScheduleRejectsMalformedJSON(t *testing.T) {
	h := New(&service.Service{Repo: repository.NewMemory(), Clock: clock.Real{}, MaxRetries: 1})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/event-1/schedule", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()
	h.Schedule(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhooksRejectsTrailingJSON(t *testing.T) {
	h := New(&service.Service{Repo: repository.NewMemory(), Clock: clock.Real{}, MaxRetries: 1})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(`{} {}`))
	req.Header.Set("X-Event-Type", "test")
	req.Header.Set("X-Source", "source")
	req.Header.Set("X-Target-URL", "https://example.com")
	rec := httptest.NewRecorder()
	h.Webhooks(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhooksReturnsConflictForDuplicateIdempotencyKey(t *testing.T) {
	h := New(&service.Service{Repo: repository.NewMemory(), Clock: clock.Real{}, MaxRetries: 1})
	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(`{"ok":true}`))
		req.Header.Set("X-Event-Type", "test")
		req.Header.Set("X-Source", "source")
		req.Header.Set("X-Target-URL", "https://example.com")
		req.Header.Set("X-Idempotency-Key", "same-key")
		rec := httptest.NewRecorder()
		h.Webhooks(rec, req)
		return rec
	}
	if status := request().Code; status != http.StatusCreated {
		t.Fatalf("first request status = %d, want %d", status, http.StatusCreated)
	}
	if status := request().Code; status != http.StatusConflict {
		t.Fatalf("duplicate request status = %d, want %d", status, http.StatusConflict)
	}
}

func TestWebhooksClassifiesReceiveErrors(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.Repository
		headers map[string]string
		want    int
	}{
		{
			name: "validation failure",
			repo: repository.NewMemory(),
			headers: map[string]string{
				"X-Event-Type": "test",
				"X-Source":     "source",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "repository failure",
			repo: failingCreateEventRepository{},
			headers: map[string]string{
				"X-Event-Type": "test",
				"X-Source":     "source",
				"X-Target-URL": "https://example.com",
			},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&service.Service{Repo: tt.repo, Clock: clock.Real{}, MaxRetries: 1})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(`{"ok":true}`))
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}
			rec := httptest.NewRecorder()
			h.Webhooks(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
