.PHONY: all build build-web build-server dev clean run test

all: build

build-web:
	cd web && pnpm install && pnpm build

build-server:
	go build -o bin/online-mermaid ./cmd/online-mermaid

build: build-web build-server

test:
	go test -v ./...
	cd web && pnpm typecheck

test-e2e: build
	cd web && pnpm test:e2e

run: build-server
	./bin/online-mermaid

clean:
	rm -rf bin build coverage web/dist internal/webui/dist
	mkdir -p internal/webui/dist
	cp internal/webui/placeholder.html internal/webui/dist/index.html 2>/dev/null || true
