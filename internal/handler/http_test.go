package handler

import (
	"bytes"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
