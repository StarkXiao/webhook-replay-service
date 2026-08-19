package service

import (
	"context"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"testing"
	"time"
)

func TestStatisticsCountsAllPages(t *testing.T) {
	repo := repository.NewMemory()
	now := time.Now()
	for i := 0; i < 501; i++ {
		if err := repo.CreateEvent(context.Background(), &domain.WebhookEvent{ID: fmt.Sprintf("event-%03d", i), Status: domain.Success, CreatedAt: now.Add(time.Duration(i) * time.Nanosecond)}); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := (&Service{Repo: repo, Clock: clock.Real{}}).Statistics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 501 || stats.Success != 501 {
		t.Fatalf("statistics = %#v", stats)
	}
}
