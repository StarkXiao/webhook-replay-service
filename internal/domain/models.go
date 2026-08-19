package domain

import "time"

type EventStatus string

const (
	Pending          EventStatus = "pending"
	Scheduled        EventStatus = "scheduled"
	Delivering       EventStatus = "delivering"
	Success          EventStatus = "success"
	Retrying         EventStatus = "retrying"
	Failed           EventStatus = "failed"
	DeadLetterStatus EventStatus = "dead_letter"
	Cancelled        EventStatus = "cancelled"
)

type WebhookEvent struct {
	ID, EventType, Source, TargetURL, IdempotencyKey string
	Headers                                          map[string]string
	Body                                             []byte
	Status                                           EventStatus
	RetryCount, MaxRetries                           int
	ReceivedAt, CreatedAt, UpdatedAt                 time.Time
	LastError                                        string
	DeletedAt                                        *time.Time
}
type DeliveryTask struct {
	ID, EventID                       string
	Status                            EventStatus
	ScheduledAt, CreatedAt, UpdatedAt time.Time
	TargetURL                         string
	Attempt                           int
	Error                             string
}
type DeliveryAttempt struct {
	ID, TaskID             string
	Attempt                int
	StatusCode             int
	ResponseSummary, Error string
	StartedAt, FinishedAt  time.Time
}
type DeadLetter struct {
	ID, EventID, Reason  string
	Attempts             []DeliveryAttempt
	CreatedAt, UpdatedAt time.Time
	Discarded            bool
}
type ReplayBatch struct {
	ID        string
	EventIDs  []string
	TargetURL string
	CreatedAt time.Time
	Status    string
}
type AuditLog struct {
	ID, Action, EntityType, EntityID, Detail string
	CreatedAt                                time.Time
}
type EventFilter struct {
	EventType, Source, Status, TargetURL, Keyword string
	From, To                                      *time.Time
	Retry, Dead                                   *bool
	Limit, Offset                                 int
	Sort                                          string
}
