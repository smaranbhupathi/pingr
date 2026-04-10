# Pingr

Uptime monitoring and status page for developers. Pingr checks your URLs on a schedule, tracks response times, opens incidents when services go down, and sends email alerts on down and recovery events.

**Live demo:** https://pingr.pages.dev

---

## Features

- HTTP uptime monitoring with configurable check intervals
- Incident tracking with downtime duration
- Response time graphs (last 24h)
- Public status page per user — share with your team or customers
- Email alerts on down and recovery
- Alert channel subscriptions — multiple emails per monitor
- Pause / resume monitors
- Per-route rate limiting (sliding window, pluggable store)
- JWT authentication with email verification and password reset

---

## Architecture

Pingr follows **hexagonal architecture** (ports and adapters). Business logic in the core has zero knowledge of HTTP, Postgres, or email — it only talks through interfaces.

```
┌─────────────────────────────────────────────────────┐
│                    Frontend (React)                  │
│                  Cloudflare Pages CDN                │
└────────────────────────┬────────────────────────────┘
                         │ HTTPS
┌────────────────────────▼────────────────────────────┐
│                   API (Go / Chi)                     │
│  Rate limiter → Auth middleware → Handlers           │
│                                                      │
│  ┌─────────────────────────────────────────────┐    │
│  │              Core (pure Go)                  │    │
│  │  services/auth   services/monitor            │    │
│  │  services/user                               │    │
│  │         ↕ interfaces only                    │    │
│  │  ports/inbound   ports/outbound              │    │
│  └─────────────────────────────────────────────┘    │
│                                                      │
│  Adapters: Postgres  │  Resend email                 │
└──────────────────────┼──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                  Worker (Go)                         │
│  Polls DB for due monitors → HTTP check              │
│  → saves result → opens/resolves incidents           │
│  → sends alerts via Resend                           │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│           PostgreSQL (Neon — AWS Singapore)          │
│  Partitioned monitor_checks table (by month)         │
└─────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, TanStack Query, Recharts |
| Backend | Go 1.24, Chi router, JWT, bcrypt |
| Database | PostgreSQL 16 (Neon serverless), golang-migrate |
| Email | Resend |
| Infrastructure | Fly.io (API + Worker), Cloudflare Pages (Frontend), Neon (DB) |
| CI/CD | GitHub Actions — test → migrate → deploy |

---

## Project Structure

```
pingr/
├── cmd/
│   ├── api/          # HTTP server entrypoint
│   └── worker/       # Background monitor checker entrypoint
├── internal/
│   ├── core/
│   │   ├── domain/         # Domain models (Monitor, User, Incident…)
│   │   ├── ports/
│   │   │   ├── inbound/    # Service interfaces (what handlers call)
│   │   │   └── outbound/   # Repository + email interfaces (what services call)
│   │   └── services/       # Business logic — no HTTP, no DB, no email
│   └── adapters/
│       ├── inbound/http/
│       │   ├── handler/    # HTTP handlers
│       │   ├── middleware/  # Auth, logger
│       │   └── ratelimit/  # Sliding window rate limiter (pluggable store)
│       └── outbound/
│           ├── postgres/   # Repository implementations
│           ├── email/      # Resend + console email senders
│           └── checker/    # HTTP monitor checker
├── migrations/       # SQL migration files (golang-migrate)
├── tests/
│   └── services/     # Unit tests — black-box, no DB needed
└── frontend/         # React SPA
```

---

## Running Locally

**Prerequisites:** Go 1.24+, Node 20+, Docker

**1. Clone and install**
```bash
git clone https://github.com/smaranbhupathi/pingr.git
cd pingr
cp .env.example .env   # fill in values
```

**2. Start Postgres**
```bash
docker compose up -d postgres
```

**3. Run migrations**
```bash
migrate -path ./migrations -database "$DATABASE_URL" up
```

**4. Start the API**
```bash
go run ./cmd/api
```

**5. Start the Worker** (separate terminal)
```bash
go run ./cmd/worker
```

**6. Start the Frontend** (separate terminal)
```bash
cd frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`, API at `http://localhost:8080`.

---

## Running Tests

```bash
go test ./tests/...
```

Tests use in-memory mocks — no database or network required. Covers auth (register, login, verify email, password reset) and monitor business rules (plan limits, interval enforcement, pause/resume, ownership).

---

## CI/CD Pipeline

Every push to `main`:

```
go test ./tests/...
       ↓ pass
golang-migrate up   (against Neon production DB)
       ↓ success
fly deploy API  ──┐
fly deploy Worker ┘  (parallel)
```

If tests fail → no deploy. If migrations fail → no deploy, old version keeps running.

---

## Key Design Decisions

**Hexagonal architecture** — core services depend on interfaces, not implementations. Swapping Postgres for another DB, or Resend for another email provider, means writing a new adapter and changing one line in `main.go`.

**Partitioned checks table** — `monitor_checks` is range-partitioned by month. As check volume grows, old partitions can be archived or dropped without touching active data.

**Pluggable rate limiter** — `ratelimit.Store` interface means the current in-memory sliding window store can be replaced with Redis (for multi-instance deployments) by implementing one interface and changing one line in `main.go`.

**Soft deletes** — monitors and alert channels have a `deleted_at` column. Historical check data and incidents are preserved after deletion.

**JWT + refresh tokens** — short-lived access tokens (15 min) with long-lived refresh tokens (7 days). Refresh is handled transparently in the Axios interceptor.

---

## Environment Variables

See `.env.example` for all required variables with descriptions.
