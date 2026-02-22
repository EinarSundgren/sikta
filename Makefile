.PONY: dev infra backend frontend migrate migration generate test build down logs setup extract dump-demo seed-demo migrate-to-graph backup-db rollback-graph eval-build

.DEFAULT_GOAL := help

BACKEND_DIR  := ./api
FRONTEND_DIR := ./web
COMPOSE_FILE := podman-compose.yaml

# Load .env if it exists (silently, so missing file is fine)
-include .env
export

# Ensure .env is always loaded for shell commands
ifneq (wildcard .env,.env)
$(info Loading .env file...)
include .env
export
endif

POSTGRES_USER     ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_DB       ?= sikta
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
DATABASE_URL      ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start infra + backend + frontend (Ctrl+C to stop all)
	@$(MAKE) infra
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Starting backend and frontend..."
	@( \
	  cd $(BACKEND_DIR) && (command -v air > /dev/null 2>&1 && air || go run ./cmd/server) \
	) & \
	( \
	  cd $(FRONTEND_DIR) && npm run dev \
	) & \
	trap 'kill %1 %2 2>/dev/null; exit 0' INT TERM; \
	wait

infra: ## Start PostgreSQL via podman compose
	podman compose -f $(COMPOSE_FILE) up -d db

backend: ## Run backend with hot reload (air) or go run
	@if command -v air > /dev/null 2>&1; then \
		cd $(BACKEND_DIR) && air; \
	else \
		cd $(BACKEND_DIR) && go run ./cmd/server; \
	fi

frontend: ## Run frontend dev server
	cd $(FRONTEND_DIR) && npm run dev

migrate: ## Apply all database migrations
	@migrate -path $(BACKEND_DIR)/sql/schema -database "$(DATABASE_URL)" up

migration: ## Create a new migration (usage: make migration name=add_something)
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migration name=add_something"; exit 1; fi
	migrate create -ext sql -dir $(BACKEND_DIR)/sql/schema -seq $(name)

generate: ## Run sqlc code generation
	cd $(BACKEND_DIR) && sqlc generate

test: ## Run all tests
	cd $(BACKEND_DIR) && go test ./...

build: ## Build all containers
	podman compose -f $(COMPOSE_FILE) build

down: ## Stop all services
	podman compose -f $(COMPOSE_FILE) down

logs: ## Stream service logs
	podman compose -f $(COMPOSE_FILE) logs -f

setup: ## Install required dev tools (air, sqlc, golang-migrate, npm deps)
	@echo "Installing Go tools..."
	go install github.com/air-verse/air@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install
	@echo ""
	@echo "Setup complete. Make sure GOPATH/bin is in your PATH:"
	@echo "  export PATH=\$$PATH:\$$(go env GOPATH)/bin"

extract: ## Run extraction on a document (usage: make extract doc=path/to/file.txt)
	@if [ -z "$(doc)" ]; then echo "Error: doc is required. Usage: make extract doc=path/to/file.txt"; exit 1; fi
	cd $(BACKEND_DIR) && go run ./cmd/extract $(doc)

dump-demo: ## Dump current database to demo/seed.sql (preserves Pride and Prejudice extraction)
	@echo "Dumping demo data to demo/seed.sql..."
	@mkdir -p demo
	@PGPASSWORD=$(POSTGRES_PASSWORD) pg_dump \
		--host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) \
		--username=$(POSTGRES_USER) --dbname=$(POSTGRES_DB) \
		--data-only --disable-triggers \
		--table=sources \
		--table=chunks \
		--table=claims \
		--table=entities \
		--table=relationships \
		--table=claim_entities \
		--table=source_references \
		--table=inconsistencies \
		--table=inconsistency_items \
		-f demo/seed.sql
	@echo "Done. Seed file: demo/seed.sql"

seed-demo: ## Load pre-extracted demo data (Pride and Prejudice) into database
	@if [ ! -f demo/seed.sql ]; then echo "Error: demo/seed.sql not found. Run 'make dump-demo' first."; exit 1; fi
	@echo "Seeding demo data from demo/seed.sql..."
	@PGPASSWORD=$(POSTGRES_PASSWORD) psql \
		--host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) \
		--username=$(POSTGRES_USER) --dbname=$(POSTGRES_DB) \
		-f demo/seed.sql
	@echo "Done. Demo data loaded."

migrate-to-graph: ## Migrate a document to graph model (usage: make migrate-to-graph doc=<source_id>)
	@if [ -z "$(doc)" ]; then \
		echo "Usage: make migrate-to-graph doc=<source_id>"; \
		echo ""; \
		echo "Available sources:"; \
		PGPASSWORD=$(POSTGRES_PASSWORD) psql \
			--host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) \
			--username=$(POSTGRES_USER) --dbname=$(POSTGRES_DB) \
			-c "SELECT id, title FROM sources;" | tail -n +2; \
		exit 1; \
	fi
	cd $(BACKEND_DIR) && go run ./cmd/migrate $(doc)

backup-db: ## Backup database before migration
	@echo "Backing up database..."
	@mkdir -p backups
	@PGPASSWORD=$(POSTGRES_PASSWORD) pg_dump \
		--host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) \
		--username=$(POSTGRES_USER) --dbname=$(POSTGRES_DB) \
		--format=plain \
		--no-owner --no-privileges \
		-f backups/sikta-backup-$$(date +%Y%m%d-%H%M%S).sql
	@echo "Database backed up to backups/"

rollback-graph: ## Rollback graph migration (delete graph tables)
	@echo "WARNING: This will delete all graph data. Continue? [y/N]"
	@read -r response; \
	if [ "$$response" = "y" ] || [ "$$response" = "Y" ]; then \
		echo "Rolling back graph migration..."; \
		PGPASSWORD=$(POSTGRES_PASSWORD) psql \
			--host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) \
			--username=$(POSTGRES_USER) --dbname=$(POSTGRES_DB) \
			-c "DROP TABLE IF EXISTS provenance, edges, nodes CASCADE;"; \
		echo "Graph tables dropped. Legacy data intact."; \
	else \
		echo "Rollback cancelled."; \
	fi

eval-build: ## Build the extraction validation CLI (sikta-eval)
	cd $(BACKEND_DIR) && go build -o ../sikta-eval ./cmd/evaluate/

eval-brf: ## Run extraction + scoring on BRF corpus (requires ANTHROPIC_API_KEY)
	@$(MAKE) eval-brf-extract
	@$(MAKE) eval-brf-score

eval-brf-extract: ## Run extraction on BRF corpus only
	@if [ -z "$(ANTHROPIC_API_KEY)" ]; then \
		echo "Error: ANTHROPIC_API_KEY not set. Please set it in .env or environment."; \
		exit 1; \
	fi
	./sikta-eval extract \
		--corpus corpora/brf \
		--prompt prompts/system/v1.txt \
		--fewshot prompts/fewshot/brf.txt \
		--output results/brf-v1.json

eval-brf-score: ## Score BRF extraction results
	./sikta-eval score \
		--result results/brf-v1.json \
		--manifest corpora/brf/manifest.json

eval-mna: ## Run extraction + scoring on M&A corpus (requires ANTHROPIC_API_KEY)
	@$(MAKE) eval-mna-extract
	@$(MAKE) eval-mna-score

eval-mna-extract: ## Run extraction on M&A corpus only
	@if [ -z "$(ANTHROPIC_API_KEY)" ]; then \
		echo "Error: ANTHROPIC_API_KEY not set. Please set it in .env or environment."; \
		exit 1; \
	fi
	./sikta-eval extract \
		--corpus corpora/mna \
		--prompt prompts/system/v1.txt \
		--fewshot prompts/fewshot/mna.txt \
		--output results/mna-v1.json

eval-mna-score: ## Score M&A extraction results
	./sikta-eval score \
		--result results/mna-v1.json \
		--manifest corpora/mna/manifest.json

eval-police: ## Run extraction + scoring on Police corpus (requires ANTHROPIC_API_KEY)
	@$(MAKE) eval-police-extract
	@$(MAKE) eval-police-score

eval-police-extract: ## Run extraction on Police corpus only
	@if [ -z "$(ANTHROPIC_API_KEY)" ]; then \
		echo "Error: ANTHROPIC_API_KEY not set. Please set it in .env or environment."; \
		exit 1; \
	fi
	./sikta-eval extract \
		--corpus corpora/police \
		--prompt prompts/system/v1.txt \
		--fewshot prompts/fewshot/police.txt \
		--output results/police-v1.json

eval-police-score: ## Score Police extraction results
	./sikta-eval score \
		--result results/police-v1.json \
		--manifest corpora/police/manifest.json

eval-all: ## Run extraction + scoring on all three corpora (requires ANTHROPIC_API_KEY)
	@echo "Running extraction validation on all corpora..."
	@$(MAKE) eval-brf
	@$(MAKE) eval-mna
	@$(MAKE) eval-police
	@echo ""
	@echo "All extractions and scoring complete!"
	@echo "Check results/*.json for extraction outputs"

eval-score-all: ## Score all existing extraction results (no API calls required)
	./sikta-eval score --result results/brf-v1.json --manifest corpora/brf/manifest.json
	./sikta-eval score --result results/mna-v1.json --manifest corpora/mna/manifest.json
	./sikta-eval score --result results/police-v1.json --manifest corpora/police/manifest.json

eval-compare: ## Compare two extraction results (usage: make eval-compare a=results/brf-v1.json b=results/brf-v2.json)
	@if [ -z "$(a)" ] || [ -z "$(b)" ]; then \
		echo "Usage: make eval-compare a=results/brf-v1.json b=results/brf-v2.json"; \
		exit 1; \
	fi
	./sikta-eval compare --a $(a) --b $(b) --manifest corpora/brf/manifest.json

