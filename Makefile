VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.1")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w -X online_mermaid/internal/config.Version=$(VERSION) -X online_mermaid/internal/config.GitCommit=$(GIT_COMMIT) -X online_mermaid/internal/config.BuildDate=$(BUILD_DATE)

.PHONY: all build build-web build-server dev clean run test test-e2e version

all: build

build-web:
	cd web && pnpm install && pnpm build

build-server:
	mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/online-mermaid ./cmd/online-mermaid

build: build-web build-server

test:
	go test -v ./...
	cd web && pnpm typecheck

test-e2e: build
	cd web && pnpm test:e2e

version: build-server
	./bin/online-mermaid version

run: build-server
	./bin/online-mermaid

clean:
	rm -rf bin build coverage web/dist internal/webui/dist
	mkdir -p internal/webui/dist
	cp internal/webui/placeholder.html internal/webui/dist/index.html 2>/dev/null || true
