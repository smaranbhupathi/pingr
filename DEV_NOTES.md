# Pingr — Dev Notes

> Single reference for commands, config, credentials, and test flows.
> Keep this file updated as the project grows.

---

## Services

| Service  | Port | Log file            |
|----------|------|---------------------|
| API      | 8080 | /tmp/pingr-api.log  |
| Worker   | —    | /tmp/pingr-worker.log |
| Frontend | 5173 | terminal output     |
| Postgres | 5432 | Docker              |

---

## Start / Stop Commands

### Start everything (3 separate terminals)

```bash
# Terminal 1 — API
cd ~/Documents/go_projects/pingr
go run ./cmd/api

# Terminal 2 — Worker
cd ~/Documents/go_projects/pingr
go run ./cmd/worker

# Terminal 3 — Frontend
cd ~/Documents/go_projects/pingr/frontend
npm run dev
```

### Start in background (single terminal)

```bash
cd ~/Documents/go_projects/pingr

# Start all
go run ./cmd/api    > /tmp/pingr-api.log    2>&1 & echo $! > /tmp/pingr-api.pid
go run ./cmd/worker > /tmp/pingr-worker.log 2>&1 & echo $! > /tmp/pingr-worker.pid
(cd frontend && npm run dev > /tmp/pingr-fe.log 2>&1 & echo $! > /tmp/pingr-fe.pid)

echo "All services started"
```

### Stop all background services

```bash
# Stop API
kill $(cat /tmp/pingr-api.pid) 2>/dev/null && rm /tmp/pingr-api.pid

# Stop Worker
kill $(cat /tmp/pingr-worker.pid) 2>/dev/null && rm /tmp/pingr-worker.pid

# Stop Frontend
kill $(cat /tmp/pingr-fe.pid) 2>/dev/null && rm /tmp/pingr-fe.pid

# Or kill by port (fallback)
lsof -ti:8080 | xargs kill -9 2>/dev/null   # API
lsof -ti:5173 | xargs kill -9 2>/dev/null   # Frontend
pkill -f "cmd/worker"                        # Worker
```

### Check what's running

```bash
lsof -i:8080 | grep LISTEN   # API
lsof -i:5173 | grep LISTEN   # Frontend
pgrep -a -f "cmd/worker"     # Worker
```

### Tail logs

```bash
tail -f /tmp/pingr-api.log      # API logs
tail -f /tmp/pingr-worker.log   # Worker logs
tail -f /tmp/pingr-fe.log       # Frontend logs (background only)
```

### Better log viewing

```bash
# Option 1 — humanlog (prettifies JSON into colored readable lines)
tail -f /tmp/pingr-api.log | humanlog
tail -f /tmp/pingr-worker.log | humanlog

# Install humanlog (if not installed)
go install github.com/humanlogio/humanlog/cmd/humanlog@latest

# Option 2 — lnav (best: both files merged, color-coded, searchable TUI)
brew install lnav
lnav /tmp/pingr-api.log /tmp/pingr-worker.log
# Inside lnav: / to search · e to jump to next error · q to quit

# Option 3 — colored terminal output (APP_ENV=dev uses tint handler)
APP_ENV=dev go run ./cmd/worker   # colored logs in terminal, console alerts only
```

---

## Credentials

> These are local/dev credentials only. Never use these in production.

### Local Postgres (Docker)
| Field    | Value  |
|----------|--------|
| Host     | localhost:5432 |
| Database | upmon  |
| Username | upmon  |
| Password | upmon  |
| Full URL | `postgres://upmon:upmon@localhost:5432/upmon?sslmode=disable` |

### MinIO (local object storage for avatar uploads)
| Field    | Value       |
|----------|-------------|
| API port | 9000        |
| Console  | http://localhost:9001 |
| Username | minioadmin  |
| Password | minioadmin  |

Start MinIO:
```bash
docker run -d --name minio \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

After starting: open http://localhost:9001 → login → create a bucket named `avatars` → set it to **public**.

Add these to your `.env` for avatar uploads locally:
```
STORAGE_ENDPOINT=http://localhost:9000
STORAGE_REGION=us-east-1
STORAGE_ACCESS_KEY_ID=minioadmin
STORAGE_SECRET_ACCESS_KEY=minioadmin
STORAGE_BUCKET=avatars
STORAGE_PUBLIC_BASE_URL=http://localhost:9000/avatars
```

### Neon (production database)
Credentials are in your `.env` file (`DATABASE_URL`).
If lost: Neon dashboard → your project → Connection Details.

---

## Database

```bash
# Start Postgres (Docker)
docker compose up -d postgres

# Stop Postgres
docker compose down

# Connect via psql
psql "postgres://upmon:upmon@localhost:5432/upmon?sslmode=disable"

# Run migrations
migrate -path ./migrations -database "$DATABASE_URL" up
migrate -path ./migrations -database "$DATABASE_URL" down 1
```

### Useful DB queries

```sql
-- All monitors
SELECT id, name, url, status, last_checked_at FROM monitors;

-- Alert channels + subscriptions
SELECT ac.id, ac.config, s.monitor_id
FROM alert_channels ac
LEFT JOIN alert_subscriptions s ON s.alert_channel_id = ac.id;

-- Recent incidents
SELECT id, monitor_id, started_at, resolved_at FROM incidents ORDER BY started_at DESC LIMIT 10;
```

---

## Environment (.env)

File location: `~/Documents/go_projects/pingr/.env`

| Key               | Notes                                              |
|-------------------|----------------------------------------------------|
| `APP_ENV`         | `dev` = console emails/alerts · `production` = Resend |
| `DATABASE_URL`    | `postgres://upmon:upmon@localhost:5432/upmon?sslmode=disable` |
| `JWT_SECRET`      | Change before deploying                            |
| `RESEND_API_KEY`  | From resend.com dashboard                          |
| `FROM_EMAIL`      | `onboarding@resend.dev` (free tier)                |
| `APP_BASE_URL`    | `http://localhost:5173` (local) · your domain (prod) |
| `WORKER_REGION`   | `us-east` (just a label, no infra dependency)      |

**Resend free tier limitation:** can only send emails to `smaranbhupathi2002@gmail.com`
(the Resend account email). To send to any address, verify a domain at resend.com/domains.

---

## End-to-End Test Flow

1. **Start testserver** (simulates a real service)
   ```bash
   go run ./cmd/testserver   # runs on http://localhost:9999
   ```

2. **Register** at `http://localhost:5173/register`
   - Use `smaranbhupathi2002@gmail.com` for real email delivery (Resend free tier)

3. **Verify email** — check inbox for link

4. **Login** → dashboard

5. **Add monitor**
   - URL: `http://localhost:9999`
   - Interval: `60s`

6. **Add alert channel** → Sidebar → Alert Channels → `smaranbhupathi2002@gmail.com`

7. **Subscribe** → click monitor → scroll to Alert Channels section → subscribe

8. **Simulate downtime** — stop the testserver (`Ctrl+C` in its terminal)
   - Worker detects DOWN within ~10s
   - DOWN email sent to Gmail

9. **Recover** — start testserver again
   - Worker detects UP within ~10s
   - RECOVERY email sent to Gmail

10. **View status page** → Sidebar → Status Page ↗

---

## Build

```bash
go build ./...              # Build all Go packages (verify no errors)
go test ./...               # Run all tests

cd frontend
npx tsc --noEmit            # TypeScript type check
npm run build               # Production build
```

---

## Project Structure (key paths)

```
cmd/
  api/          — HTTP API entry point
  worker/       — Uptime check worker entry point
  testserver/   — Tiny HTTP server for local testing (port 9999)

internal/
  core/
    domain/     — Domain types (Monitor, Incident, AlertChannel…)
    ports/
      inbound/  — Service interfaces + input types
      outbound/ — Repository + notifier interfaces
    services/   — Business logic
  adapters/
    inbound/http/   — Chi router, handlers, middleware
    outbound/
      postgres/     — All DB repositories
      email/        — Resend sender + console sender
      checker/      — HTTP checker + worker loop

frontend/src/
  api/          — Axios API clients
  components/   — Reusable UI (Button, Card, Sidebar, Footer…)
  pages/        — Route pages
  lib/          — Helpers (format.ts)

migrations/     — SQL migration files
```
