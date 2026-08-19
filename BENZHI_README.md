# Webhook Replay Service

Webhook Replay Service accepts JSON webhooks, records delivery metadata in memory, and can schedule or replay events to HTTP endpoints.

## Prerequisites

- Go 1.22 or newer
- Docker Desktop for container verification

## Commands

```bash
go build ./...
go test ./...
go run ./cmd/server
```

The server listens on `:8080` by default. Use `GET /health` to verify it, then send a JSON request to `POST /api/v1/webhooks` with `X-Event-Type`, `X-Source`, and `X-Target-URL` headers.

## Configuration

`PORT`, `MAX_RETRIES`, `INITIAL_BACKOFF_SECONDS`, `MAX_BACKOFF_SECONDS`, `WORKER_COUNT`, `REQUEST_BODY_LIMIT`, and `RATE_PER_SECOND` are optional environment variables. Invalid or non-positive integer values use documented defaults.

## Container build

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh webhook-replay-service linux/arm64
./build_benzhi_docker.sh webhook-replay-service linux/amd64
docker run -it webhook-replay-service:latest
```

The Docker image retains the Go toolchain and pre-downloads module dependencies during image creation.
