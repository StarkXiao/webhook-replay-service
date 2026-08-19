# Bug 005

## Baseline

`base_bug_005`

## Reproduction

```bash
go test ./internal/delivery -count=1 -run '^TestBug005_ClientPropagatesCancelledContext$'
```

The test cancels the caller context before sending. The baseline replaces it with a background context, allowing the request to complete instead of returning `context.Canceled`.
