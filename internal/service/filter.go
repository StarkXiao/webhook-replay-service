package service

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"sort"
	"strings"
	"time"
)

type Page struct {
	Items   []domain.WebhookEvent `json:"items"`
	Total   int                   `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasMore bool                  `json:"has_more"`
}

func NormalizeFilter(f domain.EventFilter) domain.EventFilter {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.Sort == "" {
		f.Sort = "created_at_desc"
	}
	return f
}
func FilterEvents(ctx context.Context, s *Service, f domain.EventFilter) (Page, error) {
	f = NormalizeFilter(f)
	items, total, err := s.Events(ctx, f)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset, HasMore: f.Offset+len(items) < total}, nil
}
func Match(e domain.WebhookEvent, f domain.EventFilter) bool {
	if f.EventType != "" && e.EventType != f.EventType {
		return false
	}
	if f.Source != "" && e.Source != f.Source {
		return false
	}
	if f.Status != "" && string(e.Status) != f.Status {
		return false
	}
	if f.TargetURL != "" && e.TargetURL != f.TargetURL {
		return false
	}
	if f.From != nil && e.CreatedAt.Before(*f.From) {
		return false
	}
	if f.To != nil && e.CreatedAt.After(*f.To) {
		return false
	}
	if f.Keyword != "" && !strings.Contains(strings.ToLower(string(e.Body)), strings.ToLower(f.Keyword)) {
		return false
	}
	return true
}
func SortEvents(items []domain.WebhookEvent, field string, desc bool) {
	sort.SliceStable(items, func(i, j int) bool {
		var less bool
		switch field {
		case "updated_at":
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		case "received_at":
			less = items[i].ReceivedAt.Before(items[j].ReceivedAt)
		default:
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		if desc {
			return !less
		}
		return less
	})
}
func TimeRange(from, to string) (*time.Time, *time.Time, error) {
	var a, b *time.Time
	if from != "" {
		t, e := time.Parse(time.RFC3339, from)
		if e != nil {
			return nil, nil, e
		}
		a = &t
	}
	if to != "" {
		t, e := time.Parse(time.RFC3339, to)
		if e != nil {
			return nil, nil, e
		}
		b = &t
	}
	return a, b, nil
}
