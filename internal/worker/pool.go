package worker

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/domain"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"sync"
)

type Pool struct {
	S    *service.Service
	Jobs chan domain.DeliveryTask
	wg   sync.WaitGroup
}

func New(s *service.Service, n int) *Pool {
	if n < 1 {
		n = 1
	}
	p := &Pool{S: s, Jobs: make(chan domain.DeliveryTask, n*2)}
	for i := 0; i < n; i++ {
		p.wg.Add(1)
		go p.loop()
	}
	return p
}
func (p *Pool) loop() {
	defer p.wg.Done()
	for t := range p.Jobs {
		ctx, cancel := context.WithTimeout(context.Background(), 30e9)
		_ = p.S.Deliver(ctx, &t)
		cancel()
	}
}
func (p *Pool) Submit(ctx context.Context, t domain.DeliveryTask) bool {
	select {
	case p.Jobs <- t:
		return true
	case <-ctx.Done():
		return false
	}
}
func (p *Pool) Close() { close(p.Jobs); p.wg.Wait() }
