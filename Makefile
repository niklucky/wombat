BINARY := wombat
MAIN := ./cmd/wombat
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
EXT ?= $(if $(filter windows,$(GOOS)),.exe,)
BINARY_NAME ?= $(BINARY)
OUTPUT := dist/$(BINARY_NAME)$(EXT)

.PHONY: all build build-all test run-tui run-tray clean tidy

all: build

build:
	@mkdir -p dist
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(OUTPUT) $(MAIN)

build-all:
	@mkdir -p dist
	$(MAKE) build GOOS=darwin GOARCH=amd64 BINARY_NAME=wombat-darwin-amd64
	$(MAKE) build GOOS=darwin GOARCH=arm64 BINARY_NAME=wombat-darwin-arm64
	$(MAKE) build GOOS=linux GOARCH=amd64 BINARY_NAME=wombat-linux-amd64
	$(MAKE) build GOOS=linux GOARCH=arm64 BINARY_NAME=wombat-linux-arm64
	$(MAKE) build GOOS=windows GOARCH=amd64 BINARY_NAME=wombat-windows-amd64 EXT=.exe

test:
	go test ./...

run-tui:
	go run -ldflags "$(LDFLAGS)" $(MAIN) tui

run-tray:
	go run -ldflags "$(LDFLAGS)" $(MAIN) tray-daemon

clean:
	rm -rf dist
	rm -f $(BINARY)

tidy:
	go mod tidy
