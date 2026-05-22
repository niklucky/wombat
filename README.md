# Wombat

A cross-platform SSH helper built with Go. Features a Terminal UI (TUI), system tray application, desktop notifications, and an optional GUI.

## Features

- **TUI** — Navigate and manage SSH hosts, keys, and tunnels with a keyboard-driven Bubble Tea interface.
- **System Tray** — Keep Wombat running in the background with quick access via the tray icon.
- **Notifications** — Get notified about connection events and tunnel status changes.
- **GUI** *(optional)* — A Fyne-based desktop interface, available when built with the `gui` tag.
- **CLI** — Direct commands for scripting and automation.

## Quick Start

```bash
# Build the standard binary (TUI + tray + CLI)
make build

# Run the TUI
./wombat tui

# Run the system tray app
./wombat tray

# List configured hosts
./wombat cli list

# Build with GUI support
make build-gui
./wombat-gui gui
```

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
│   └── gui/           # Fyne GUI (tagged build)
├── assets/            # Icons and static assets
└── scripts/           # Build scripts
```

## Platform Support

- macOS
- Linux
- Windows

## License

MIT
