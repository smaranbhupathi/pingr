.PHONY: start stop start-api stop-api start-worker stop-worker start-frontend stop-frontend \
        status logs logs-api logs-worker restart \
        db db-stop migrate migrate-down build tidy test

# ── Start / Stop ─────────────────────────────────────────────────────────────

start: start-api start-worker start-frontend
	@echo "All services started. Run 'make status' to verify."

stop: stop-api stop-worker stop-frontend
	@echo "All services stopped."

restart: stop start

start-api:
	@go run ./cmd/api > /tmp/pingr-api.log 2>&1 & echo $$! > /tmp/pingr-api.pid
	@echo "API started (PID $$(cat /tmp/pingr-api.pid)) → logs: /tmp/pingr-api.log"

stop-api:
	@if [ -f /tmp/pingr-api.pid ]; then \
		kill $$(cat /tmp/pingr-api.pid) 2>/dev/null && rm /tmp/pingr-api.pid && echo "API stopped"; \
	else \
		lsof -ti:8080 | xargs kill -9 2>/dev/null && echo "API stopped (fallback)" || echo "API not running"; \
	fi

start-worker:
	@go run ./cmd/worker > /tmp/pingr-worker.log 2>&1 & echo $$! > /tmp/pingr-worker.pid
	@echo "Worker started (PID $$(cat /tmp/pingr-worker.pid)) → logs: /tmp/pingr-worker.log"

stop-worker:
	@if [ -f /tmp/pingr-worker.pid ]; then \
		kill $$(cat /tmp/pingr-worker.pid) 2>/dev/null && rm /tmp/pingr-worker.pid && echo "Worker stopped"; \
	else \
		pkill -f "cmd/worker" 2>/dev/null && echo "Worker stopped (fallback)" || echo "Worker not running"; \
	fi

start-frontend:
	@cd frontend && npm run dev > /tmp/pingr-fe.log 2>&1 & echo $$! > /tmp/pingr-fe.pid
	@echo "Frontend started (PID $$(cat /tmp/pingr-fe.pid)) → http://localhost:5173"

stop-frontend:
	@if [ -f /tmp/pingr-fe.pid ]; then \
		kill $$(cat /tmp/pingr-fe.pid) 2>/dev/null && rm /tmp/pingr-fe.pid && echo "Frontend stopped"; \
	else \
		lsof -ti:5173 | xargs kill -9 2>/dev/null && echo "Frontend stopped (fallback)" || echo "Frontend not running"; \
	fi

start-testserver:
	@go run ./cmd/testserver > /tmp/pingr-test.log 2>&1 & echo $$! > /tmp/pingr-test.pid
	@echo "Test server started on http://localhost:9999 (PID $$(cat /tmp/pingr-test.pid))"

stop-testserver:
	@if [ -f /tmp/pingr-test.pid ]; then \
		kill $$(cat /tmp/pingr-test.pid) 2>/dev/null && rm /tmp/pingr-test.pid && echo "Test server stopped"; \
	else \
		lsof -ti:9999 | xargs kill -9 2>/dev/null || echo "Test server not running"; \
	fi

# ── Status / Logs ─────────────────────────────────────────────────────────────

status:
	@echo "=== Service Status ==="
	@lsof -i:8080 | grep LISTEN > /dev/null 2>&1 && echo "API        ✓ running (port 8080)" || echo "API        ✗ stopped"
	@lsof -i:5173 | grep LISTEN > /dev/null 2>&1 && echo "Frontend   ✓ running (port 5173)" || echo "Frontend   ✗ stopped"
	@pgrep -f "cmd/worker" > /dev/null 2>&1   && echo "Worker     ✓ running"              || echo "Worker     ✗ stopped"
	@lsof -i:9999 | grep LISTEN > /dev/null 2>&1 && echo "Testserver ✓ running (port 9999)" || echo "Testserver ✗ stopped"
	@docker ps --filter name=postgres --format "Postgres   ✓ running (port 5432)" 2>/dev/null || true

logs:
	@tail -f /tmp/pingr-api.log /tmp/pingr-worker.log 2>/dev/null

logs-api:
	@tail -f /tmp/pingr-api.log

logs-worker:
	@tail -f /tmp/pingr-worker.log

# ── Database ──────────────────────────────────────────────────────────────────

db:
	docker compose up -d postgres

db-stop:
	docker compose down

migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

# ── Build / Test ──────────────────────────────────────────────────────────────

build-api:
	go build -o bin/api ./cmd/api

build-worker:
	go build -o bin/worker ./cmd/worker

build: build-api build-worker
	@echo "Binaries written to bin/"

tidy:
	go mod tidy

test:
	go test ./...
