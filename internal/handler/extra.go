package handler

import (
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"net/http"
)

func (h *Handler) Statistics(w http.ResponseWriter, r *http.Request) {
	_, total, _ := h.S.Events(r.Context(), domain.EventFilter{Limit: 1})
	write(w, 200, map[string]int{"events": total})
}
func (h *Handler) DeadLetters(w http.ResponseWriter, r *http.Request) {
	d, _ := h.S.Repo.ListDeadLetters(r.Context())
	write(w, 200, d)
}
