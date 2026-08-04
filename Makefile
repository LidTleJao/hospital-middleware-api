COMPOSE      := docker compose
COMPOSE_DEV  := docker compose -f docker-compose.yml -f docker-compose.dev.yml
MIGRATE_IMG  := migrate/migrate:v4.18.2

# Loaded so migrate-* targets can reach Postgres from the host.
-include .env
export

DB_URL ?= postgres://$(or $(POSTGRES_USER),hospital):$(or $(POSTGRES_PASSWORD),hospital)@localhost:$(or $(POSTGRES_PORT),5432)/$(or $(POSTGRES_DB),hospital)?sslmode=disable

.DEFAULT_GOAL := help

## ---------- environment

.PHONY: env
env: ## Create .env from .env.example if it does not exist yet
	@test -f .env || (cp .env.example .env && echo "created .env from .env.example")

## ---------- docker

.PHONY: up
up: env ## Build and start the full stack (nginx + api + postgres + migrate)
	$(COMPOSE) up --build -d
	@echo "API is available at http://localhost:$(or $(HTTP_PORT),8080)"

.PHONY: dev
dev: env ## Start the stack with hot reload (air), streaming logs
	$(COMPOSE_DEV) up --build

.PHONY: down
down: ## Stop the stack, keep the database volume
	$(COMPOSE) down

.PHONY: clean
clean: ## Stop the stack and delete the database volume
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Follow logs of every service
	$(COMPOSE) logs -f

.PHONY: logs-api
logs-api: ## Follow logs of the api service only
	$(COMPOSE) logs -f api

.PHONY: ps
ps: ## Show container status
	$(COMPOSE) ps

.PHONY: psql
psql: ## Open a psql shell inside the postgres container
	$(COMPOSE) exec postgres psql -U $(or $(POSTGRES_USER),hospital) -d $(or $(POSTGRES_DB),hospital)

## ---------- migrations

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(COMPOSE) run --rm migrate

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	docker run --rm -v $(PWD)/migrations:/migrations --network host $(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" down 1

.PHONY: migrate-force
migrate-force: ## Clear a dirty migration state: make migrate-force version=1
	docker run --rm -v $(PWD)/migrations:/migrations --network host $(MIGRATE_IMG) \
		-path=/migrations -database "$(DB_URL)" force $(version)

.PHONY: migrate-create
migrate-create: ## Scaffold a migration pair: make migrate-create name=add_patients
	docker run --rm -v $(PWD)/migrations:/migrations $(MIGRATE_IMG) \
		create -ext sql -dir /migrations -seq $(name)

## ---------- go

.PHONY: run
run: ## Run the API on the host (needs a reachable Postgres)
	go run ./cmd/api

.PHONY: build
build: ## Compile the binary into ./bin/api
	go build -trimpath -o bin/api ./cmd/api

.PHONY: test
test: ## Run all tests
	go test ./... -count=1

.PHONY: test-cover
test-cover: ## Run tests and write an HTML coverage report
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "report: coverage.html"

.PHONY: fmt
fmt: ## Format every Go file
	go fmt ./...

.PHONY: vet
vet: ## Run the built-in static checks
	go vet ./...

.PHONY: tidy
tidy: ## Sync go.mod / go.sum with the imports actually used
	go mod tidy

.PHONY: check
check: fmt vet test ## Format, vet and test in one go

## ---------- help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'
