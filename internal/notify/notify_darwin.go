//go:build darwin

package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var iconPath string

func init() {
	// Extract icon to a temp location so terminal-notifier can reference it
	tmpIcon := filepath.Join(os.TempDir(), "wombat-notification-icon.png")
	if _, err := os.Stat(tmpIcon); os.IsNotExist(err) {
		_ = os.WriteFile(tmpIcon, iconData, 0644)
	}
	iconPath = tmpIcon
}

// Notify sends a desktop notification.
func Notify(title, message string) error {
	return notify(title, message, false)
}

// Alert sends a desktop alert with sound.
func Alert(title, message string) error {
	return notify(title, message, true)
}

func notify(title, message string, urgent bool) error {
	// Only use terminal-notifier if it is installed on the system.
	// Bundled/extracted copies are blocked by Gatekeeper, so we avoid them.
	if tn, err := exec.LookPath("terminal-notifier"); err == nil {
		args := []string{
			"-title", title,
			"-message", message,
			"-group", "Wombat",
			"-appIcon", iconPath,
		}
		if urgent {
			args = append(args, "-sound", "default")
		}
		cmd := exec.Command(tn, args...)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to osascript — always works, no icon
	return osascriptNotify(title, message)
}

func osascriptNotify(title, message string) error {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}
	script := fmt.Sprintf("display notification %q with title %q sound name \"default\"", message, title)
	return exec.Command(osa, "-e", script).Run()
}
