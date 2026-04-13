# Pingr

Open-source uptime monitoring and public status pages.

Add a URL, pick a check interval, and Pingr will watch it around the clock — alerting you on Discord, Slack, or email the moment it goes down, and again when it recovers. Share a public status page with your team or customers so everyone always knows what's happening.

**Live demo:** https://getpingr.com

---

## Features

- **HTTP uptime monitoring** — configurable check intervals (30 s → 24 h)
- **Public status page** — one URL to share with your users, updated in real time
- **Incident management** — create, update, and resolve incidents with a full timeline
- **Component groups** — group monitors into logical services (API, Database, CDN…)
- **Component status** — Operational / Degraded / Partial Outage / Major Outage / Under Maintenance
- **Multi-channel alerts** — Email, Slack, and Discord; enable/disable per channel
- **Import / export alert channels** — CSV and JSON, with conflict resolution
- **Response time graphs** — 24-hour latency chart per monitor
- **90-day uptime bars** — visual history on the status page
- **Pause / resume monitors** — stop checks without deleting history
- **JWT auth** — email verification, password reset, refresh tokens
- **Rate limiting** — sliding-window per route, pluggable store (in-memory → Redis)
- **Dark mode** — full light/dark theme support

---

## Architecture

Pingr uses **hexagonal architecture** (ports and adapters). Core business logic has zero knowledge of HTTP, Postgres, or email — it only talks through interfaces. Swapping any adapter means writing a new struct and changing one line in `main.go`.

```
┌──────────────────────────────────────────┐
│           Frontend  (React + Vite)        │
│           Served from CDN                 │
└──────────────────┬───────────────────────┘
                   │ HTTPS / JSON
┌──────────────────▼───────────────────────┐
│           API Server  (Go / Chi)          │
│  Rate limiter → Auth JWT → Handlers       │
│                                           │
│  ┌───────────────────────────────────┐   │
│  │         Core  (pure Go)           │   │
│  │  services/auth                    │   │
│  │  services/monitor                 │   │
│  │  services/user                    │   │
│  │       ↕  interfaces only          │   │
│  │  ports/inbound  ports/outbound    │   │
│  └───────────────────────────────────┘   │
│                                           │
│  Adapters:  postgres │ email │ webhook    │
└──────────────────┬───────────────────────┘
                   │
┌──────────────────▼───────────────────────┐
│        Worker  (Go background process)    │
│  Polls DB for due monitors               │
│  → runs HTTP check                       │
│  → saves result                          │
│  → opens / resolves incidents            │
│  → sends alerts                          │
└──────────────────┬───────────────────────┘
                   │
┌──────────────────▼───────────────────────┐
│              PostgreSQL                   │
└──────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, TanStack Query, Recharts |
| Backend | Go 1.24, Chi router, pgx v5, JWT, bcrypt |
| Database | PostgreSQL 16, golang-migrate |
| Email | Resend (or console logger in dev — no account needed) |
| Alerts | Resend (email), Slack Incoming Webhooks, Discord Webhooks |
| Infrastructure | Any Docker-capable host (Railway, Fly.io, Render, VPS…) |

---

## Project Structure

```
pingr/
├── cmd/
│   ├── api/            # HTTP server entry point
│   ├── worker/         # Background monitor checker entry point
│   └── testserver/     # Tiny HTTP server for local testing (port 9999)
├── internal/
│   ├── core/
│   │   ├── domain/         # Domain models: Monitor, User, Incident, Component…
│   │   ├── ports/
│   │   │   ├── inbound/    # Service interfaces + input types (what handlers call)
│   │   │   └── outbound/   # Repository + notifier interfaces (what services call)
│   │   └── services/       # Business logic — no HTTP, no DB, no email
│   └── adapters/
│       ├── inbound/http/
│       │   ├── handler/    # HTTP request handlers
│       │   ├── middleware/  # Auth (JWT), request logger
│       │   └── ratelimit/  # Sliding-window rate limiter with pluggable store
│       └── outbound/
│           ├── postgres/   # All repository implementations
│           ├── email/      # Resend sender + console sender (dev)
│           ├── webhook/    # Slack + Discord notifiers
│           ├── checker/    # HTTP monitor checker + worker loop
│           └── storage/    # S3-compatible object storage (avatars)
├── migrations/         # SQL migration files (golang-migrate, numbered)
├── tests/
│   └── services/       # Unit tests — in-memory mocks, no DB required
├── frontend/           # React SPA
│   └── src/
│       ├── api/        # Typed API clients (Axios)
│       ├── components/ # Reusable UI components
│       ├── pages/      # Route-level page components
│       └── lib/        # Utility helpers
├── docker-compose.yml  # Local Postgres (port 5432)
├── .env.example        # Copy to .env and fill in values
└── config.yaml         # Feature flags (email/slack/discord alerts)
```

---

## Running Locally

### Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://www.docker.com/) (for local Postgres)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

Install golang-migrate:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 1. Clone the repo

```bash
git clone https://github.com/smaranbhupathi/pingr.git
cd pingr
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Open `.env` and fill in the required values. For a minimal local setup you only need:

```env
DATABASE_URL=postgres://upmon:upmon@localhost:5432/upmon?sslmode=disable
JWT_SECRET=any-long-random-string-here
APP_BASE_URL=http://localhost:5173
APP_ENV=dev
ALLOWED_ORIGIN=*
```

> **Email in dev mode:** when `APP_ENV=dev` and `RESEND_API_KEY` is not set, all emails and alerts are printed to the terminal instead of being sent. No Resend account needed to get started.

### 3. Start Postgres

```bash
docker compose up -d postgres
```

This starts a Postgres 16 container on port 5432 with database `upmon`, user `upmon`, password `upmon`.

### 4. Run database migrations

```bash
migrate -path ./migrations -database "postgres://upmon:upmon@localhost:5432/upmon?sslmode=disable" up
```

### 5. Start the API server

```bash
go run ./cmd/api
# → listening on http://localhost:8080
```

### 6. Start the worker (separate terminal)

```bash
go run ./cmd/worker
# → polls for due monitors every 10 s
```

### 7. Start the frontend (separate terminal)

```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173
```

Open [http://localhost:5173](http://localhost:5173), register an account, and add your first monitor.

### 8. (Optional) Run a local test server

If you don't have a service to monitor, Pingr ships a tiny HTTP server for testing:

```bash
go run ./cmd/testserver
# → http://localhost:9999 — returns 200 OK
```

Add `http://localhost:9999` as a monitor. Stop it to simulate downtime.

---

## Running Tests

```bash
go test ./tests/...
```

Tests use in-memory mocks — no database or network required. Covers:
- Auth: register, login, email verification, password reset
- Monitor rules: plan limits, interval enforcement, pause/resume, ownership checks

---

## Configuration

Feature flags live in `config.yaml` at the project root. You can toggle alert channel types without rebuilding:

```yaml
features:
  email_alerts:   true
  slack_alerts:   true
  discord_alerts: true

monitoring:
  worker_tick_seconds: 10
```

---

## Environment Variables Reference

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `JWT_SECRET` | Yes | Secret for signing JWTs — generate with `openssl rand -hex 32` |
| `APP_BASE_URL` | Yes | Public URL of the frontend (used in email links) |
| `APP_ENV` | No | `dev` = console emails/alerts · `production` = Resend (default: `dev`) |
| `PORT` | No | HTTP port for the API server (default: `8080`) |
| `ALLOWED_ORIGIN` | No | CORS origin — use `*` in dev, exact frontend URL in prod |
| `WORKER_REGION` | No | Label for this worker instance (default: `us-east`) |
| `RESEND_API_KEY` | No | [Resend](https://resend.com) API key — leave unset to use console logging |
| `FROM_EMAIL` | No | Sender address for emails (must be verified in Resend) |
| `STORAGE_ENDPOINT` | No | S3-compatible endpoint for avatar uploads (Cloudflare R2, MinIO, AWS S3) |
| `STORAGE_ACCESS_KEY_ID` | No | Storage access key ID |
| `STORAGE_SECRET_ACCESS_KEY` | No | Storage secret key |
| `STORAGE_BUCKET` | No | Bucket name |
| `STORAGE_REGION` | No | Storage region (use `auto` for Cloudflare R2) |
| `STORAGE_PUBLIC_BASE_URL` | No | Public base URL for uploaded files |

See `.env.example` for a ready-to-copy template.

---

## Deploying

Pingr is two Go binaries (`api` and `worker`) plus a static frontend — deployable anywhere.

**Database:** Any PostgreSQL 16+ instance. [Neon](https://neon.tech) (serverless, free tier) works well. Point `DATABASE_URL` at your instance and run migrations.

**API + Worker:** Any Docker or Go-capable host (Railway, Fly.io, Render, a plain VPS). Build with:
```bash
go build -o bin/api    ./cmd/api
go build -o bin/worker ./cmd/worker
```

**Frontend:** `npm run build` produces a static `dist/` folder. Deploy to Cloudflare Pages, Vercel, Netlify, or any static host. Set `VITE_API_URL` to your API's public URL.

**Migrations on deploy:** Run `migrate up` against your production database before starting the new binary. A failed migration will leave the old binary running unchanged.

---

## Key Design Decisions

**Hexagonal architecture** — core services have no imports from `net/http`, `pgx`, or any email library. All I/O goes through interfaces defined in `ports/`. This makes the business logic independently testable with in-memory fakes.

**Separate API and Worker** — the API serves requests; the worker runs the check loop. They share the same database but are deployed independently. You can scale them separately or run multiple workers in different regions.

**Soft deletes** — monitors and alert channels have a `deleted_at` column. Historical check data and incidents are preserved after deletion so uptime graphs stay accurate.

**Pluggable rate limiter** — `ratelimit.Store` is an interface. The current in-memory sliding window can be replaced with Redis (for multi-instance deployments) by implementing the interface and changing one line in `main.go`.

**JWT + refresh tokens** — short-lived access tokens (15 min) with long-lived refresh tokens (7 days). Refresh is handled transparently in the Axios interceptor — users stay logged in without re-entering credentials.

**No-account dev mode** — when `RESEND_API_KEY` is absent, all emails and alert notifications are printed to stdout. You can develop and test the full flow without any external service accounts.

---

## Contributing

Pull requests are welcome. For larger changes, open an issue first to discuss the approach.

```bash
# Run tests before submitting
go test ./tests/...

# Type-check the frontend
cd frontend && npx tsc --noEmit
```

Please keep PRs focused — one feature or fix per PR. Match the existing code style (no linter config yet, just follow what's already there).

---

## License

MIT
