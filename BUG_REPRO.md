# Bug 002

## Baseline

`base_bug_002`

## Reproduction

```bash
go test ./internal/delivery -count=1 -run '^TestBug002_NewClientCanSendRequest$'
```

The test creates a client through `delivery.New` and sends a request. The baseline panics because the client's HTTP implementation is nil.
