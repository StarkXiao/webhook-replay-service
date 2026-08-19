package service

import (
	"sync"
	"testing"
	"time"
)

func TestBug001_GateAcquireConcurrentAccess(t *testing.T) {
	gate := NewGate()
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				gate.Acquire("event-1", time.Millisecond)
			}
		}()
	}
	wg.Wait()
}
