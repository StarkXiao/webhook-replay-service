package service

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
)

type Statistics struct {
	Total, Pending, Scheduled, Delivering, Success, Retrying, Failed, DeadLetter, Cancelled int `json:"-"`
}

func (s Statistics) Map() map[string]int {
	return map[string]int{"total": s.Total, "pending": s.Pending, "scheduled": s.Scheduled, "delivering": s.Delivering, "success": s.Success, "retrying": s.Retrying, "failed": s.Failed, "dead_letter": s.DeadLetter, "cancelled": s.Cancelled}
}
func (s *Service) Statistics(ctx context.Context) (Statistics, error) {
	const pageSize = 500
	out := Statistics{}
	for offset := 0; ; offset += pageSize {
		items, total, err := s.Events(ctx, domain.EventFilter{Limit: pageSize, Offset: offset})
		if err != nil {
			return out, err
		}
		out.Total = total
		for _, e := range items {
			switch e.Status {
			case domain.Pending:
				out.Pending++
			case domain.Scheduled:
				out.Scheduled++
			case domain.Delivering:
				out.Delivering++
			case domain.Success:
				out.Success++
			case domain.Retrying:
				out.Retrying++
			case domain.Failed:
				out.Failed++
			case domain.DeadLetterStatus:
				out.DeadLetter++
			case domain.Cancelled:
				out.Cancelled++
			}
		}
		if offset+len(items) >= total {
			return out, nil
		}
	}
}
