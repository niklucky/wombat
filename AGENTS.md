# Agent Notes for Wombat

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

## Release Process

Releases are automated via GitHub Actions when a `v*.*.*` tag is pushed.

### Required Secrets

| Secret | Purpose |
|--------|---------|
| `SNAPCRAFT_STORE_CREDENTIALS` | Publish to the Snap Store. Generate with `snapcraft export-login <file>` and paste the contents. |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Push formula updates to `niklucky/homebrew-tap`. Must have `repo` scope. |

### Distribution Channels

- **GitHub Releases** — Pre-built binaries for macOS, Linux, and Windows.
- **Homebrew** — Personal tap at `niklucky/homebrew-tap`. Install with `brew install niklucky/tap/wombat`.
- **Snap Store** — Published as a *classic* snap. First upload requires manual approval from Canonical for classic confinement. Install with `snap install wombat --classic`.
