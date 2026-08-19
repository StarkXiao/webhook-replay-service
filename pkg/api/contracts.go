package api

import (
	"encoding/json"
	"time"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
type Response struct {
	Data      any    `json:"data,omitempty"`
	Error     *Error `json:"error,omitempty"`
	Meta      Meta   `json:"meta,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}
type Meta struct {
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Total      int    `json:"total,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}
type WebhookRequest struct {
	EventID        string            `json:"event_id"`
	EventType      string            `json:"event_type"`
	Source         string            `json:"source"`
	TargetURL      string            `json:"target_url"`
	Payload        json.RawMessage   `json:"payload"`
	Headers        map[string]string `json:"headers,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Signature      string            `json:"signature,omitempty"`
}
type WebhookBatchRequest struct {
	Events []WebhookRequest `json:"events"`
}
type ScheduleRequest struct {
	DeliverAt    *time.Time `json:"deliver_at,omitempty"`
	DelaySeconds int        `json:"delay_seconds,omitempty"`
}
type ReplayRequest struct {
	EventIDs       []string   `json:"event_ids,omitempty"`
	TargetURL      string     `json:"target_url,omitempty"`
	DeliverAt      *time.Time `json:"deliver_at,omitempty"`
	DelaySeconds   int        `json:"delay_seconds,omitempty"`
	MaxConcurrency int        `json:"max_concurrency,omitempty"`
}
type EventResponse struct {
	ID         string    `json:"id"`
	EventType  string    `json:"event_type"`
	Source     string    `json:"source"`
	TargetURL  string    `json:"target_url"`
	Status     string    `json:"status"`
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	ReceivedAt time.Time `json:"received_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastError  string    `json:"last_error,omitempty"`
}
type TaskResponse struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	Status      string    `json:"status"`
	TargetURL   string    `json:"target_url"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Attempt     int       `json:"attempt"`
	Error       string    `json:"error,omitempty"`
}
type AttemptResponse struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	Attempt         int       `json:"attempt"`
	StatusCode      int       `json:"status_code,omitempty"`
	ResponseSummary string    `json:"response_summary,omitempty"`
	Error           string    `json:"error,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
}
type DeadLetterResponse struct {
	ID        string            `json:"id"`
	EventID   string            `json:"event_id"`
	Reason    string            `json:"reason"`
	Discarded bool              `json:"discarded"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Attempts  []AttemptResponse `json:"attempts"`
}
type BatchResponse struct {
	BatchID   string   `json:"batch_id"`
	Status    string   `json:"status"`
	Requested int      `json:"requested"`
	Accepted  int      `json:"accepted"`
	Failed    int      `json:"failed"`
	TaskIDs   []string `json:"task_ids"`
}
type HealthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}
type StatisticsResponse struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Scheduled  int `json:"scheduled"`
	Delivering int `json:"delivering"`
	Success    int `json:"success"`
	Retrying   int `json:"retrying"`
	Failed     int `json:"failed"`
	DeadLetter int `json:"dead_letter"`
	Cancelled  int `json:"cancelled"`
}
type AuditResponse struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Detail     string    `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

func (r WebhookRequest) IsBatchable() bool {
	return r.EventType != "" && r.Source != "" && len(r.Payload) > 0
}
func (r ScheduleRequest) HasSchedule() bool { return r.DeliverAt != nil || r.DelaySeconds > 0 }
func (r ReplayRequest) IsDelayed() bool     { return r.DeliverAt != nil || r.DelaySeconds > 0 }
func EmptyResponse(requestID string) Response {
	return Response{Data: map[string]any{}, RequestID: requestID}
}
func ErrorResponse(requestID, code, message string) Response {
	return Response{Error: &Error{Code: code, Message: message}, RequestID: requestID}
}
