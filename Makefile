SHELL := /bin/bash
BIN := $(CURDIR)/.bin

GOLANGCI_LINT_VERSION := v2.14.0
GITLEAKS_VERSION := v8.30.1
LEFTHOOK_VERSION := v2.1.17
GOVULNCHECK_VERSION := v1.8.0
DEADCODE_VERSION := v0.51.0
SQLC_VERSION := v1.31.1

WITH_API_ENV := cd api && set -a && . ./.env && set +a

.PHONY: setup tools app-deps hooks env migrate neon-check dev dev-api dev-app check check-api \
	check-app check-repo test test-api test-app build gen e2e

setup: tools app-deps env hooks

tools:
	GOBIN=$(BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	GOBIN=$(BIN) go install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION)
	GOBIN=$(BIN) go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	GOBIN=$(BIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	GOBIN=$(BIN) go install golang.org/x/tools/cmd/deadcode@$(DEADCODE_VERSION)
	GOBIN=$(BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

app-deps:
	cd app && npm ci && npx playwright install chromium

env:
	go -C api run ./cmd/devenv -- ..

hooks:
	$(BIN)/lefthook install

migrate:
	$(WITH_API_ENV) && go run ./cmd/migrate up

neon-check:
	cd api && go run ./cmd/dbcheck .env.production

dev: migrate
	$(MAKE) -j2 dev-api dev-app

dev-api:
	$(WITH_API_ENV) && go run ./cmd/server

dev-app:
	cd app && npm run dev

check: check-repo check-api check-app

check-repo:
	go -C api run ./cmd/nocomment
	$(BIN)/gitleaks git --no-banner --redact .

check-api:
	cd api && $(BIN)/golangci-lint fmt --diff ./...
	cd api && $(BIN)/golangci-lint run ./...
	cd api && $(BIN)/govulncheck ./...
	cd api && unused="$$($(BIN)/deadcode -test ./...)" && test -z "$$unused" || { echo "$$unused"; exit 1; }

check-app:
	cd app && npm run check
	cd app && npm run lint
	cd app && npm audit --audit-level=moderate

test: test-api test-app

test-api:
	$(WITH_API_ENV) && TEST_DATABASE_URL="$${TEST_DATABASE_URL:-$$DATABASE_MIGRATE_URL}" go test -race ./...

test-app:
	cd app && npm test

gen:
	cd api && $(BIN)/sqlc generate
	cd app && npm run gen:types

build:
	cd api && go build -trimpath -o $(BIN)/trustip-api ./cmd/server
	cd app && npm run build

e2e: migrate
	cd app && npx playwright test
