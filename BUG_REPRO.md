# Bug 004

## Baseline

`base_bug_004`

## Reproduction

```bash
go test ./internal/service -count=1 -run '^TestBug004_ReceivePreservesDuplicateError$'
```

The test submits the same idempotency key twice and checks `errors.Is`. The baseline converts the duplicate error to text, so the original sentinel cannot be identified.
