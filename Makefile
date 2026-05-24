BINARY := wombat
MAIN := ./cmd/wombat

.PHONY: all build test run-tui run-tray clean tidy

all: build

build:
	go build -o $(BINARY) $(MAIN)

test:
	go test ./...

run-tui:
	go run $(MAIN) tui

run-tray:
	go run $(MAIN) tray

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy
