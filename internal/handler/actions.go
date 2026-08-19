package handler

import (
	"encoding/json"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"io"
	"net/http"
	"strings"
	"time"
)

type replayBody struct {
	EventIDs       []string   `json:"event_ids"`
	TargetURL      string     `json:"target_url"`
	DeliverAt      *time.Time `json:"deliver_at"`
	DelaySeconds   int        `json:"delay_seconds"`
	MaxConcurrency int        `json:"max_concurrency"`
}

func (h *Handler) Replay(w http.ResponseWriter, r *http.Request) {
	var body replayBody
	if err := decodeJSON(r, &body); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"error": "invalid replay JSON"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	eventID := parts[len(parts)-2]
	at := scheduleTime(body.DeliverAt, body.DelaySeconds)
	task, err := h.S.Replay(r.Context(), eventID, body.TargetURL, at)
	if err != nil {
		write(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusCreated, task)
}

func (h *Handler) ReplayBatch(w http.ResponseWriter, r *http.Request) {
	var body replayBody
	if err := decodeJSON(r, &body); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"error": "invalid replay JSON"})
		return
	}
	at := scheduleTime(body.DeliverAt, body.DelaySeconds)
	result, err := h.S.ReplayBatch(r.Context(), service.ReplayRequest{EventIDs: body.EventIDs, TargetURL: body.TargetURL, At: &at, MaxConcurrency: body.MaxConcurrency})
	if err != nil {
		write(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusCreated, result)
}

func (h *Handler) TaskAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	taskID, action := parts[len(parts)-2], parts[len(parts)-1]
	var err error
	switch action {
	case "retry":
		err = h.S.RetryTask(r.Context(), taskID)
	case "cancel":
		err = h.S.CancelTask(r.Context(), taskID)
	default:
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		write(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusOK, map[string]string{"status": action + "ed"})
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func scheduleTime(at *time.Time, delay int) time.Time {
	if at != nil {
		return *at
	}
	return time.Now().UTC().Add(time.Duration(delay) * time.Second)
}
