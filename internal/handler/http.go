package handler

import (
	"encoding/json"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct{ S *service.Service }

func New(s *service.Service) *Handler { return &Handler{S: s} }
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": v, "status": status})
}
func (h *Handler) Webhooks(w http.ResponseWriter, r *http.Request) {
	var b json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		write(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	e, err := h.S.Receive(r.Context(), service.ReceiveInput{Body: b, EventType: r.Header.Get("X-Event-Type"), Source: r.Header.Get("X-Source"), TargetURL: r.Header.Get("X-Target-URL"), IdempotencyKey: r.Header.Get("X-Idempotency-Key")})
	if err != nil {
		write(w, 400, map[string]string{"error": err.Error()})
		return
	}
	write(w, 201, e)
}
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	e, total, err := h.S.Events(r.Context(), domain.EventFilter{EventType: q.Get("event_type"), Source: q.Get("source"), Status: q.Get("status"), TargetURL: q.Get("target_url"), Keyword: q.Get("keyword"), Limit: limit, Offset: offset})
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	write(w, 200, map[string]any{"items": e, "pagination": map[string]int{"limit": limit, "offset": offset, "total": total}})
}
func (h *Handler) Event(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	e, err := h.S.Repo.GetEvent(r.Context(), id)
	if err != nil {
		write(w, 404, map[string]string{"error": "not found"})
		return
	}
	write(w, 200, e)
}
func (h *Handler) Schedule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	eid := parts[len(parts)-2]
	var x struct {
		At           time.Time `json:"at"`
		DelaySeconds int       `json:"delay_seconds"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&x); err != nil {
			write(w, http.StatusBadRequest, map[string]string{"error": "invalid schedule JSON"})
			return
		}
	}
	if x.At.IsZero() {
		x.At = time.Now().UTC().Add(time.Duration(x.DelaySeconds) * time.Second)
	}
	t, err := h.S.Schedule(r.Context(), eid, x.At)
	if err != nil {
		write(w, 400, map[string]string{"error": err.Error()})
		return
	}
	write(w, 201, t)
}
