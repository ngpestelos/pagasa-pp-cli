.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -o bin/pagasa-pp-cli$(BIN_EXT) ./cmd/pagasa-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/pagasa-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/pagasa-pp-mcp$(BIN_EXT) ./cmd/pagasa-pp-mcp

install-mcp:
	go install ./cmd/pagasa-pp-mcp

build-all: build build-mcp
