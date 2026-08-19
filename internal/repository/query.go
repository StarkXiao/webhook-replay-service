package repository

import (
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"strings"
	"time"
)

type Query struct {
	Where         []string
	Args          []any
	Order         string
	Limit, Offset int
}

func NewEventQuery(f domain.EventFilter) Query {
	q := Query{Order: "created_at DESC", Limit: f.Limit, Offset: f.Offset}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	q.Equal("event_type", f.EventType)
	q.Equal("source", f.Source)
	q.Equal("status", f.Status)
	q.Equal("target_url", f.TargetURL)
	if f.From != nil {
		q.After("created_at", *f.From)
	}
	if f.To != nil {
		q.Before("created_at", *f.To)
	}
	if f.Keyword != "" {
		q.Contains("body", f.Keyword)
	}
	return q
}
func (q *Query) Equal(column, value string) {
	if value == "" {
		return
	}
	q.Args = append(q.Args, value)
	q.Where = append(q.Where, fmt.Sprintf("%s = $%d", column, len(q.Args)))
}
func (q *Query) After(column string, value time.Time) {
	q.Args = append(q.Args, value)
	q.Where = append(q.Where, fmt.Sprintf("%s >= $%d", column, len(q.Args)))
}
func (q *Query) Before(column string, value time.Time) {
	q.Args = append(q.Args, value)
	q.Where = append(q.Where, fmt.Sprintf("%s <= $%d", column, len(q.Args)))
}
func (q *Query) Contains(column, value string) {
	q.Args = append(q.Args, "%"+value+"%")
	q.Where = append(q.Where, fmt.Sprintf("%s::text ILIKE $%d", column, len(q.Args)))
}
func (q Query) SQL(base string) (string, []any) {
	if len(q.Where) > 0 {
		base += " WHERE " + strings.Join(q.Where, " AND ")
	}
	if q.Order != "" {
		base += " ORDER BY " + q.Order
	}
	q.Args = append(q.Args, q.Limit, q.Offset)
	base += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(q.Args)-1, len(q.Args))
	return base, q.Args
}
func Placeholders(n, start int) string {
	p := make([]string, n)
	for i := range p {
		p[i] = fmt.Sprintf("$%d", i+start)
	}
	return strings.Join(p, ",")
}
func NullableTime(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
func NullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func ArchiveSQL(table string) string {
	return "UPDATE " + table + " SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
}
func LockDueTasksSQL(limit int) string {
	return fmt.Sprintf("SELECT id FROM delivery_tasks WHERE status IN ('pending','scheduled','retrying') AND scheduled_at <= NOW() ORDER BY scheduled_at FOR UPDATE SKIP LOCKED LIMIT %d", limit)
}
