# Agent Notes for Wombat

## Attention

Before every session make backup copies of:
* ~/.config/wombat/config.json -> /tmp/TIMESTAMP-wombat_config.json
* ~/.ssh/config.d/wombat -> /tmp/TIMESTAMP-wombat_ssh_config

## Build & Development

- Use `/usr/local/go/bin/go` if `go` is not on your PATH.
- Run `go mod tidy` after adding or removing dependencies.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep UI code isolated in `internal/tui` and `internal/tray`.
- Place all business logic and models in `internal/core`.
- Platform-specific behavior belongs in `internal/platform`.

## Adding Dependencies

- CLI: `github.com/spf13/cobra`
- TUI: `github.com/charmbracelet/bubbletea`, `lipgloss`, `bubbles`
- Tray: `github.com/gogpu/systray` (zero CGO)
- Notifications: `github.com/gen2brain/beeep`
- SSH: `golang.org/x/crypto/ssh`

## Testing Proof-of-Life

1. `make build` should produce a runnable `wombat` binary.
2. `wombat tui` should show a host list (or "No hosts configured").
3. `wombat tray-daemon` should show a tray icon with a menu.
