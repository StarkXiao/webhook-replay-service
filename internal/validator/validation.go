package validator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
type Errors []FieldError

func (e Errors) Error() string {
	parts := make([]string, len(e))
	for i, x := range e {
		parts[i] = x.Field + ": " + x.Message
	}
	return strings.Join(parts, "; ")
}
func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return Errors{{field, "is required"}}
	}
	return nil
}
func JSONBody(body []byte) error {
	if len(body) == 0 {
		return Errors{{"body", "is required"}}
	}
	if !json.Valid(body) {
		return Errors{{"body", "must be valid JSON"}}
	}
	return nil
}
func URL(field, value string) error {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
		return Errors{{field, "must be an HTTP URL"}}
	}
	return nil
}
func Future(field string, t time.Time, now time.Time) error {
	if t.Before(now) {
		return Errors{{field, "must be in the future"}}
	}
	return nil
}
func Limit(value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("value must be between %d and %d", min, max)
	}
	return nil
}
func All(errs ...error) error {
	var out Errors
	for _, err := range errs {
		if err != nil {
			if list, ok := err.(Errors); ok {
				out = append(out, list...)
			} else {
				out = append(out, FieldError{"request", err.Error()})
			}
		}
	}
	if len(out) > 0 {
		return out
	}
	return nil
}
