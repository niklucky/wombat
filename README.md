# Wombat

![Wombat](assets/app-icon-128.png)

A cross-platform SSH helper built with Go. Features a Terminal UI (TUI), system tray application, and desktop notifications.

## Features

- **TUI** — Navigate and manage SSH hosts, keys, and tunnels with a keyboard-driven Bubble Tea interface.
- **System Tray** — Keep Wombat running in the background with quick access via the tray icon.
- **Notifications** — Get notified about connection events and tunnel status changes.
- **CLI** — Direct commands for scripting and automation.

## Quick Start

For compiled binaries, see the [Releases](https://github.com/niklucky/wombat/releases) page.
If you want to build from source:

```bash
# Build the standard binary (TUI + tray + CLI)
make build

# Run the TUI
./wombat

# List configured hosts
./wombat list

```

## CLI Commands

### Host Management

| Command | Description |
|---------|-------------|
| `wombat list` | List configured SSH hosts |
| `wombat add-host <name>` | Add a new host |
| `wombat remove-host <name>` | Remove a host |
| `wombat connect <name>` | Connect to a host via SSH |
| `wombat test <name>` | Test TCP connectivity to a host |
| `wombat import-ssh-config` | Import hosts from `~/.ssh/config` |

**`add-host` flags:** `--address` (required), `--user` (required), `--port` (default: 22), `--key`

### Tunnel Management

| Command | Description |
|---------|-------------|
| `wombat tunnel-list` | List all tunnels (shows active/inactive status) |
| `wombat add-tunnel <name>` | Add a new port-forwarding tunnel |
| `wombat remove-tunnel <name>` | Remove a tunnel |
| `wombat tunnel-start <name>` | Start a tunnel in the background |
| `wombat tunnel-stop <name>` | Stop a running tunnel |

**`add-tunnel` flags:** `--host` (required), `--local-port` (required), `--remote-port` (required), `--remote-host` (default: localhost)

### Application

| Command | Description |
|---------|-------------|
| `wombat` | Launch the TUI (and optional tray daemon) |
| `wombat tray-daemon` | Launch the system tray app |
| `wombat version` | Print version information |

## Project Structure

```
wombat/
├── cmd/wombat/        # Entry point
├── internal/
│   ├── core/          # Business logic & config
│   ├── sshutil/       # SSH client & agent helpers
│   ├── platform/      # OS-specific utilities
│   ├── notify/        # Desktop notifications
│   ├── tui/           # Bubble Tea TUI
│   ├── tray/          # System tray app
│   └── tunnelmgr/     # Tunnel management
├── assets/            # Icons and static assets
└── scripts/           # Build scripts
```

## Platform Support

- macOS
- Linux
- Windows

## License

MIT
