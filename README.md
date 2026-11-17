# WebhookLab

Self-hostable webhook capture and inspection tool. Create a unique URL, point any HTTP client at it, and inspect every incoming request in a live dashboard.

Stack: Go (chi + pgx) backend, React + TypeScript (Vite) frontend, PostgreSQL.

## Features
- Create webhook endpoints with unique URLs (`/w/{slug}`).
- List, view details, copy URL, enable/disable, delete endpoints (cascades events).
- Auto-expiration after a configurable TTL (default 7 days); expired endpoints stop accepting requests.
- Per-endpoint event cap (default 1000) to bound storage.
- Captures method, headers (JSONB), body (raw bytes), query params (JSONB), content type, source IP, timestamp, size, path.
- Live updates via Server-Sent Events.
- Replay captured requests to any destination URL with editable method, headers, and body.
- Replay history with response status, headers, body, and duration.
- SSRF protection: private/loopback/link-local/metadata destinations are blocked by default.
- Single `docker compose up` to run.

## Architecture
See design notes in the repo history. High level:
- `backend/` — Go HTTP API + capture handler + expiry worker.
- `frontend/` — React SPA.
- `postgres` — endpoints and events tables; JSONB for headers/query params; BYTEA for raw body.

## Prerequisites
- Docker + Docker Compose (recommended), OR
- Go 1.22+, Node 20+, PostgreSQL 16+ for local dev.

## Run with Docker Compose
```bash
docker compose up --build
```
- Dashboard: http://localhost:3000
- API: http://localhost:8080

## Run locally (dev)
Backend:
```bash
cd backend
export DATABASE_URL="postgres://webhooklab:webhooklab@localhost:5432/webhooklab?sslmode=disable"
go run ./cmd/server
```
Frontend:
```bash
cd frontend
npm install
npm run dev
```
Dashboard at http://localhost:5173 (proxies /api and /w to backend).

## Configuration (env)
| Var | Default | Description |
|---|---|---|
| DATABASE_URL | required | Postgres connection string |
| HTTP_ADDR | :8080 | Listen address |
| WEBHOOKLAB_BASE_URL | http://localhost:8080 | Base URL used to build webhook URLs |
| WEBHOOKLAB_DEFAULT_TTL | 168h | Default endpoint TTL (0 = never) |
| WEBHOOKLAB_MAX_BODY_BYTES | 1048576 | Max captured body size |
| WEBHOOKLAB_CLEANUP_INTERVAL | 1h | Expiry worker interval |
| WEBHOOKLAB_EXPIRED_RETENTION | 720h | Keep expired endpoints this long before purge |
| WEBHOOKLAB_EVENT_CAP | 1000 | Max events per endpoint (0 = unlimited) |
| WEBHOOKLAB_CORS_ORIGINS | * | Comma-separated allowed origins |
| WEBHOOKLAB_TRUSTED_PROXIES | (none) | (reserved) |
| WEBHOOKLAB_REPLAY_TIMEOUT | 10s | Outbound replay timeout |
| WEBHOOKLAB_REPLAY_MAX_RESPONSE_BYTES | 1048576 | Max stored replay response body |
| WEBHOOKLAB_REPLAY_ALLOW_PRIVATE | false | Allow replays to private/loopback hosts (dev only; metadata is always blocked) |

## API (top-level)
Management under `/api/v1/endpoints`; capture under `/w/{slug}`; health at `/healthz`, `/readyz`.

## License
MIT
```
Need postgres local setup maybe. README enough.

Now there may be compile issues due to service import of repository captured event; ok.

Let me rebuild backend after all.<|end|><|start|><|start|>assistant to=shell<|channel|>analysis <|constrain|>xml<|message|><rfl:parameter name="command">
cd /home/boss/Desktop/projects/webhooklab/backend && go vet ./... 2>&1 | tail -20 && go build ./... 2>&1 | tail