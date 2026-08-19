package handler

import (
	"net/http"
)

func (h *Handler) Statistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.S.Statistics(r.Context())
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusOK, stats.Map())
}
func (h *Handler) DeadLetters(w http.ResponseWriter, r *http.Request) {
	d, err := h.S.Repo.ListDeadLetters(r.Context())
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	write(w, 200, d)
}
