package domain

import "fmt"

func (s EventStatus) Valid() bool {
	switch s {
	case Pending, Scheduled, Delivering, Success, Retrying, Failed, DeadLetterStatus, Cancelled:
		return true
	}
	return false
}
func (s EventStatus) Terminal() bool { return s == Success || s == DeadLetterStatus || s == Cancelled }
func (s EventStatus) CanDeliver() bool {
	return s == Pending || s == Scheduled || s == Retrying || s == Failed
}
func (e *WebhookEvent) Transition(next EventStatus) error {
	if !next.Valid() {
		return fmt.Errorf("invalid status %q", next)
	}
	if e.Status == Cancelled || e.Status == DeadLetterStatus {
		return fmt.Errorf("event is terminal")
	}
	e.Status = next
	return nil
}
func (t *DeliveryTask) Transition(next EventStatus) error {
	if !next.Valid() {
		return fmt.Errorf("invalid status %q", next)
	}
	if t.Status == Success || t.Status == Cancelled {
		return fmt.Errorf("task is terminal")
	}
	t.Status = next
	return nil
}
func (d *DeadLetter) Restore() { d.Discarded = false }
func (d *DeadLetter) Discard() { d.Discarded = true }
func Statuses() []EventStatus {
	return []EventStatus{Pending, Scheduled, Delivering, Success, Retrying, Failed, DeadLetterStatus, Cancelled}
}
