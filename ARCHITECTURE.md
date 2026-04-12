# Architecture

This document explains the key design decisions behind Pingr — not just *what* was built, but *why*.

---

## Hexagonal Architecture (Ports and Adapters)

The core of Pingr — auth, monitor rules, incident management — has **zero imports** from `net/http`, `pgx`, `resend`, or any external library. It only talks to the outside world through interfaces defined in `ports/`.

```
ports/inbound/   — interfaces the HTTP handlers call into (UserService, MonitorService…)
ports/outbound/  — interfaces the services call into (MonitorRepository, EmailSender…)
```

**Why:**

- The business logic is testable with simple in-memory fakes. No database, no HTTP server, no email account needed to run the test suite.
- Swapping an adapter (e.g. Postgres → another DB, Resend → SendGrid) means writing one new struct and changing one line in `main.go`. Nothing in the core changes.
- The boundary between core and infrastructure is explicit and enforced by the compiler. It's impossible to accidentally call `pgx` from a service.

---

## Two Separate Processes: API and Worker

The API server and the monitor checker are separate binaries deployed independently.

```
cmd/api/    — serves HTTP requests
cmd/worker/ — runs the uptime check loop
```

**Why:**

- They have different scaling characteristics. The API scales with request traffic; the worker scales with monitor count. Separating them means you can add more workers without adding more API instances.
- A crash in the worker doesn't take down the API, and vice versa.
- Running multiple workers in different regions (AWS us-east, ap-southeast…) requires zero code changes — each worker just has a different `WORKER_REGION` env var and only picks up monitors tagged for that region.

---

## Worker Design: Poll, Don't Subscribe

The worker polls the database on a tight loop rather than using a message queue or pub/sub system.

```go
// Every N seconds: fetch monitors where last_checked_at <= now - interval
monitors, _ := repo.GetDue(ctx, region)
```

**Why:**

- No additional infrastructure (no Redis, no Kafka, no RabbitMQ) needed to run the project.
- The database is the source of truth for scheduling. If a worker crashes mid-batch, the next tick picks up the missed monitors naturally — no message acknowledgement or dead-letter queue needed.
- For the scale Pingr targets (thousands of monitors, not millions), polling is fast enough and operationally simpler.

**Trade-off:** At very high monitor counts, polling the DB every 10 seconds becomes a bottleneck. The fix at that point is a message queue — the worker interface already supports it without changing the core.

---

## Component Status vs Monitor Status

Pingr tracks two distinct statuses per monitor:

| Status | Who sets it | What it means |
|---|---|---|
| `monitor.status` | Worker only | Internal: is the HTTP check passing? (`up` / `down`) |
| `monitor.component_status` | Worker + operator | Public: what should the status page show? |

**Why:**

- A monitor can be `down` internally (HTTP check failing) but an operator may want to show `under_maintenance` on the public status page rather than `major_outage`.
- Decoupling them means the worker can auto-set component status on outage/recovery, while the operator retains full control to override it via incidents.
- The status page never exposes raw `up`/`down` — it only shows the human-readable component status.

---

## Incident System

Incidents are the public-facing communication layer. They are separate from outage events.

```
OutageEvent  — internal, created by worker, used for uptime math
Incident     — public, shown on status page, has a timeline of updates
```

The worker auto-creates an incident when a monitor goes down and auto-resolves it on recovery. Operators can also create incidents manually (planned maintenance, partial degradation) independent of any monitor check.

**Why:**

- Not every outage needs a public incident, and not every incident is caused by a monitor check (planned maintenance, partial degradation that doesn't trip the check threshold).
- Keeping them separate means the uptime calculation (based on outage events) is never affected by how operators choose to communicate incidents.

---

## Authentication: JWT + Refresh Tokens

Access tokens are short-lived (15 minutes). Refresh tokens are long-lived (7 days) and stored in the database.

**Why:**

- Short access token lifetime limits the window of exposure if a token is intercepted.
- The Axios interceptor handles refresh transparently — the user never sees a login prompt during normal use.
- Refresh tokens in the database means they can be revoked server-side (logout, password reset, suspicious activity).

---

## Rate Limiting: Pluggable Store

The rate limiter uses a `Store` interface with an in-memory sliding window implementation.

```go
type Store interface {
    Allow(key string, limit int, window time.Duration) bool
    Close()
}
```

**Why:**

- The in-memory store works perfectly for a single-instance deployment (which covers most use cases).
- When horizontal scaling requires shared state, replace it with a Redis store by implementing the same interface and changing one line in `main.go`. No handler or middleware code changes.

---

## Soft Deletes

Monitors and alert channels have a `deleted_at` column instead of being hard-deleted.

**Why:**

- Deleting a monitor would orphan its historical check data and break uptime graphs.
- Deleting an alert channel would orphan its subscriptions and silently drop alert history.
- Soft deletes preserve the full audit trail while hiding deleted records from the application.

---

## Connection Pool Design

Both the API and Worker use `pgxpool` with `MinConns=0`.

**Why:**

- `MinConns=0` means no connections are opened on startup. Connections are acquired on first use and released when idle.
- This prevents the "too many clients" error during rolling deploys, where old and new processes briefly overlap.
- It also allows the database compute (Neon serverless) to suspend when idle rather than staying awake to serve pre-opened idle connections.

---

## No-Account Dev Mode

When `RESEND_API_KEY` is not set, all emails and alert notifications are printed to stdout via console adapters.

```go
// email/console.go — implements the same EmailSender interface as the Resend adapter
func (s *ConsoleSender) Send(...) error {
    slog.Info("email (console)", "to", to, "subject", subject)
    return nil
}
```

**Why:**

- A new contributor can clone the repo, run `docker compose up -d`, and have a fully working local environment in under 5 minutes without signing up for any external service.
- The console adapters implement the same interfaces as the real adapters — the core and handlers never know the difference.
