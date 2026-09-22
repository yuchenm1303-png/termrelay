.PHONY: build generate generate-ent test test-unit test-integration test-e2e

VERSION ?= $(shell ./scripts/resolve-version.sh)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

generate-ent:
	go generate ./ent

build: generate-ent
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate: generate-ent
	go generate ./cmd/server

test: generate-ent
	go test ./...
	golangci-lint run ./...

test-unit: generate-ent
	go test -tags=unit ./...

test-integration: generate-ent
	go test -tags=integration ./...

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...
