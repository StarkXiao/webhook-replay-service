package service

import (
	"sync"
	"testing"
	"time"
)

func TestBug001_GateAcquireConcurrentAccess(t *testing.T) {
	gate := NewGate()
	var wg sync.WaitGroup
	results := make(chan bool, 64)
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- gate.Acquire("event-1", time.Second)
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	for acquired := range results {
		if acquired {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful acquires = %d, want 1", successes)
	}

	if gate.Acquire("event-1", time.Second) {
		t.Fatal("unexpired key was acquired twice")
	}
}

func TestGateExpiredKeyCanBeAcquiredAgain(t *testing.T) {
	gate := NewGate()
	if !gate.Acquire("event-1", time.Millisecond) {
		t.Fatal("initial acquire failed")
	}

	time.Sleep(5 * time.Millisecond)
	if !gate.Acquire("event-1", time.Second) {
		t.Fatal("expired key was not acquired again")
	}
}

func TestGateConcurrentAcquireReleaseAndCleanup(t *testing.T) {
	gate := NewGate()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				gate.Acquire("event-1", time.Millisecond)
				gate.Release("event-1")
				gate.Cleanup(time.Now())
			}
		}()
	}
	wg.Wait()
}
