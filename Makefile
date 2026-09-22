.PHONY: up down build test lint build-all build-web logs clean migrate

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Build a single service: make build svc=auth-service
build:
	docker compose build $(svc)

# Build all services
build-all:
	docker compose build

# Build web frontend (placeholder - add frontend service when ready)
build-web:
	@echo "Web build placeholder - add frontend service"

# Run all tests
test:
	go work sync
	go test ./services/shared/...

# Lint all Go code
lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./services/...

# View logs
logs:
	docker compose logs -f

# Clean up everything
clean:
	docker compose down -v --rmi local
	rm -rf /tmp/warungos-*

# Run Postgres migrations
migrate:
	docker exec -i warungos-postgres psql -U warungos -d warungos < migrations/postgres/001_init.up.sql

# Seed MongoDB
seed:
	docker exec -i warungos-mongo mongosh warungos < migrations/mongo/seed.js

# Sync Go workspace
sync:
	go work sync

# Tidy all modules
tidy:
	@for dir in services/shared services/api-gateway services/auth-service services/menu-service services/order-service services/payment-service services/inventory-service; do \
		echo "Tidying $$dir..."; \
		cd $$dir && go mod tidy && cd ../..; \
	done

# Quick health check
health:
	@echo "=== Service Health ==="
	@curl -sf http://localhost:8080/health || echo "api-gateway: DOWN"
	@curl -sf http://localhost:8081/health || echo "auth-service: DOWN"
	@curl -sf http://localhost:8082/health || echo "menu-service: DOWN"
	@curl -sf http://localhost:8083/health || echo "order-service: DOWN"
	@curl -sf http://localhost:8084/health || echo "payment-service: DOWN"
	@curl -sf http://localhost:8085/health || echo "inventory-service: DOWN"
