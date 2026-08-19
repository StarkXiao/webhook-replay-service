package handler

import (
	"encoding/json"
	"github.com/StarkXiao/webhook-replay-service/internal/middleware"
	"net/http"
)

type Envelope struct {
	Data      any            `json:"data,omitempty"`
	Error     *ErrorBody     `json:"error,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func respond(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data, Meta: meta, RequestID: requestID(r)})
}
func fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Error: &ErrorBody{Code: code, Message: message}, RequestID: requestID(r)})
}
func requestID(r *http.Request) string {
	if v, ok := r.Context().Value(middleware.RequestID).(string); ok {
		return v
	}
	return r.Header.Get("X-Request-ID")
}
