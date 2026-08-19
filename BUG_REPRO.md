# Bug 003

## Baseline

`base_bug_003`

## Reproduction

```bash
go test ./internal/repository -count=1 -run '^TestBug003_AttemptsReturnsIndependentSlice$'
```

The test changes one returned attempt and reads the history again. The baseline exposes its internal slice, so the caller's change contaminates stored data.
