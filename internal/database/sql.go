package database

import (
	"fmt"
	"sort"
	"strings"
)

type Statement struct {
	Name string
	SQL  string
	Args []any
}
type Transaction interface {
	Exec(string, ...any) error
	Commit() error
	Rollback() error
}

func Insert(table string, values map[string]any) Statement {
	keys := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for k, v := range values {
		keys = append(keys, k)
		args = append(args, v)
	}
	sort.Strings(keys)
	args = args[:0]
	for _, k := range keys {
		args = append(args, values[k])
	}
	holders := make([]string, len(keys))
	for i := range holders {
		holders[i] = fmt.Sprintf("$%d", i+1)
	}
	return Statement{Name: "insert_" + table, SQL: "INSERT INTO " + table + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(holders, ",") + ")", Args: args}
}
func Update(table string, values map[string]any, where string, whereArgs ...any) Statement {
	parts := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	i := 1
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := values[k]
		parts = append(parts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	for _, v := range whereArgs {
		args = append(args, v)
	}
	return Statement{Name: "update_" + table, SQL: "UPDATE " + table + " SET " + strings.Join(parts, ",") + " WHERE " + where, Args: args}
}
func Delete(table, where string, args ...any) Statement {
	return Statement{Name: "delete_" + table, SQL: "DELETE FROM " + table + " WHERE " + where, Args: args}
}
func Select(table string, columns []string, where, order string, limit int) Statement {
	q := "SELECT " + strings.Join(columns, ",") + " FROM " + table
	if where != "" {
		q += " WHERE " + where
	}
	if order != "" {
		q += " ORDER BY " + order
	}
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	return Statement{Name: "select_" + table, SQL: q}
}
func EventColumns() []string {
	return []string{"id", "event_type", "source", "target_url", "idempotency_key", "headers", "body", "status", "retry_count", "max_retries", "received_at", "created_at", "updated_at", "last_error"}
}
func TaskColumns() []string {
	return []string{"id", "event_id", "status", "scheduled_at", "target_url", "attempt", "error", "created_at", "updated_at"}
}
func AttemptColumns() []string {
	return []string{"id", "task_id", "attempt", "status_code", "response_summary", "error", "started_at", "finished_at"}
}
func DeadLetterColumns() []string {
	return []string{"id", "event_id", "reason", "discarded", "created_at", "updated_at"}
}
func AuditColumns() []string {
	return []string{"id", "action", "entity_type", "entity_id", "detail", "created_at"}
}
func IsConstraintError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "violates unique constraint"))
}
func IsTransientError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadlock"))
}
