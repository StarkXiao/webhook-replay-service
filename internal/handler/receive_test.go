package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
)

// newTestHandler builds a handler backed by a fresh in-memory repository.
func newTestHandler() *Handler {
	return New(&service.Service{Repo: repository.NewMemory(), Clock: clock.Real{}, MaxRetries: 1})
}

func postWebhooks(h *Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.Webhooks(rec, req)
	return rec
}

func TestWebhooksDuplicateIdempotencyKeyReturns409(t *testing.T) {
	h := newTestHandler()
	headers := map[string]string{
		"X-Event-Type":      "invoice.paid",
		"X-Source":          "billing",
		"X-Target-URL":      "https://example.com",
		"X-Idempotency-Key": "duplicate-key",
	}
	body := `{"ok":true}`

	if rec := postWebhooks(h, body, headers); rec.Code != http.StatusCreated {
		t.Fatalf("first POST status = %d, want %d", rec.Code, http.StatusCreated)
	}
	rec := postWebhooks(h, body, headers)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate POST status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if env.Error == nil || env.Error.Code != "duplicate_event" {
		t.Fatalf("error code = %v, want %q", env.Error, "duplicate_event")
	}
}

func TestWebhooksMissingMetadataReturns400ValidationCode(t *testing.T) {
	h := newTestHandler()
	// No X-Event-Type / X-Source / X-Target-URL → validator.Errors path.
	rec := postWebhooks(h, `{"ok":true}`, map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if env.Error == nil || env.Error.Code != "validation_failed" {
		t.Fatalf("error code = %v, want %q", env.Error, "validation_failed")
	}
}

// TestClassifyReceiveErrorOrdering pins that a duplicate is not misclassified
// as a validation error (the validation branch runs first and must not shadow
// it) and that a %w-wrapped duplicate still classifies as conflict, proving
// the error chain is restored end-to-end.
func TestClassifyReceiveErrorOrdering(t *testing.T) {
	if status, code := classifyReceiveError(repository.ErrDuplicate); status != http.StatusConflict || code != "duplicate_event" {
		t.Fatalf("ErrDuplicate classified as (%d, %q), want (%d, %q)", status, code, http.StatusConflict, "duplicate_event")
	}

	// Real wrapped-duplicate path: drive Receive twice with a configured service.
	s := &service.Service{Repo: repository.NewMemory(), Clock: clock.Real{}}
	in := service.ReceiveInput{Body: []byte(`{"ok":true}`), EventType: "invoice.paid", Source: "billing", TargetURL: "https://example.com", IdempotencyKey: "dup"}
	if _, err := s.Receive(nil, in); err != nil {
		t.Fatalf("first receive: %v", err)
	}
	_, err := s.Receive(nil, in)
	if err == nil {
		t.Fatal("second receive: expected duplicate error, got nil")
	}
	if status, code := classifyReceiveError(err); status != http.StatusConflict || code != "duplicate_event" {
		t.Fatalf("wrapped duplicate classified as (%d, %q), want (%d, %q)", status, code, http.StatusConflict, "duplicate_event")
	}
}
