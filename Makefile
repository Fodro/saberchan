.PHONY: ci backend-test backend-build backend-install-deps backend-mocks frontend-install \
	frontend-check frontend-build ensure-frontend-env local-up local-down local-logs local-ps help

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
COMPOSE_LOCAL := docker compose -f $(ROOT)docker-compose.local.yaml --env-file $(ROOT).env.local
GOBIN := $(shell go env GOPATH)/bin

ensure-frontend-env:
	@test -f $(ROOT)frontend/.env || cp $(ROOT)frontend/.env.dist $(ROOT)frontend/.env

help:
	@echo "Targets:"
	@echo "  make ci                   Run full local CI gate"
	@echo "  make backend-install-deps Install global Go tools (mockgen, etc.)"
	@echo "  make backend-test         go test ./..."
	@echo "  make backend-build        go build backend"
	@echo "  make backend-mocks        regenerate uber/gomock mocks"
	@echo "  make frontend-install     npm ci"
	@echo "  make frontend-check       svelte-check / lint types"
	@echo "  make frontend-build       production frontend build"
	@echo "  make local-up             start full local stack (compose.local)"
	@echo "  make local-down           stop local stack"
	@echo "  make local-logs           follow local stack logs"
	@echo "  make local-ps             show local stack status"

backend-install-deps:
	@echo "Installing Go tools into $(GOBIN)"
	go install go.uber.org/mock/mockgen@v0.6.0
	@command -v mockgen >/dev/null 2>&1 || { \
		echo "mockgen installed to $(GOBIN)/mockgen — add \$$(go env GOPATH)/bin to PATH"; \
	}
	@echo "backend-install-deps ok"

backend-test:
	cd $(ROOT)backend && go test ./...

backend-build:
	cd $(ROOT)backend && go build -o /tmp/saberchan-api ./main.go

# Requires mockgen from: make backend-install-deps
backend-mocks: backend-install-deps
	cd $(ROOT)backend && PATH="$(GOBIN):$$PATH" go generate ./internal/database ./internal/file ./internal/captcha

frontend-install:
	cd $(ROOT)frontend && npm ci

frontend-check: ensure-frontend-env
	cd $(ROOT)frontend && npm run check

frontend-build: ensure-frontend-env
	cd $(ROOT)frontend && npm run build

# Install once, then typecheck/build/test. Requires network on first frontend-install.
ci: frontend-install backend-test backend-build frontend-check frontend-build
	@echo "ci ok"

local-up:
	@test -f $(ROOT).env.local || (echo "Missing .env.local — copy from .env.local.dist" && exit 1)
	# One image at a time: parallel Go+Vite builds OOM Colima at 2GiB (adapter-node SIGKILL).
	COMPOSE_PARALLEL_LIMIT=1 $(COMPOSE_LOCAL) up -d --build

local-down:
	@test -f $(ROOT).env.local || (echo "Missing .env.local — copy from .env.local.dist" && exit 1)
	$(COMPOSE_LOCAL) down

local-logs:
	@test -f $(ROOT).env.local || (echo "Missing .env.local — copy from .env.local.dist" && exit 1)
	$(COMPOSE_LOCAL) logs -f

local-ps:
	@test -f $(ROOT).env.local || (echo "Missing .env.local — copy from .env.local.dist" && exit 1)
	$(COMPOSE_LOCAL) ps
