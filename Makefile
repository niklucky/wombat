BINARY := wombat
MAIN := ./cmd/wombat

.PHONY: all build build-gui run-tui run-tray run-gui clean tidy

all: build

build:
	go build -o $(BINARY) $(MAIN)

build-gui:
	go build -tags gui -o $(BINARY)-gui $(MAIN)

run-tui:
	go run $(MAIN) tui

run-tray:
	go run $(MAIN) tray

run-gui:
	go run -tags gui $(MAIN) gui

clean:
	rm -f $(BINARY) $(BINARY)-gui

tidy:
	go mod tidy
