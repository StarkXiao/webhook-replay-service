# Bug 001

## Baseline

`base_bug_001`

## Reproduction

```bash
go test -race ./internal/service -count=1 -run '^TestBug001_GateAcquireConcurrentAccess$'
```

The test starts many goroutines that call `Gate.Acquire` for the same key. The baseline reports a data race because the shared active map is accessed without synchronization.
