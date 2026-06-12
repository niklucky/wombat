package tray

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gogpu/systray"
	"github.com/niklucky/wombat/assets"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

// RunWithTunnels creates and runs the system tray UI for managing configured tunnels.
// It uses cfg to build the tray menu and tooltip, sets the tray icon from assets when available,
// writes a tray PID file, and starts a background ticker that refreshes the menu every 3 seconds
// to keep tunnel state in sync. It runs the tray event loop and logs and alerts on errors.
// When the event loop exits, the tray PID file is removed.
func RunWithTunnels(cfg core.Config) {
	tray := systray.New()
	if assets.TrayIcon != nil {
		tray.SetIcon(assets.TrayIcon)
	}

	var refresh func()
	refresh = func() {
		tray.SetMenu(buildMenu(tray, cfg, refresh))
		updateTooltip(tray, cfg)
	}

	refresh()
	// NOTE: tray.OnClick does not fire on macOS when a menu is attached
	// because NSStatusItem suppresses the button action in favor of showing
	// the menu. This is a platform limitation of gogpu/systray on macOS.
	tray.Show()

	_ = tunnelmgr.WriteTrayPidFile()

	// Refresh tray menu periodically so tunnel state stays in sync
	// when tunnels are started/stopped outside the tray.
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			refresh()
		}
	}()

	if err := tray.Run(); err != nil {
		log.Printf("tray.Run failed: %v", err)
		_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failed"), err))
	}
	_ = tunnelmgr.RemoveTrayPidFile()
}

func buildMenu(tray *systray.SystemTray, cfg core.Config, refresh func()) *systray.Menu {
	menu := systray.NewMenu()

	for _, t := range cfg.Tunnels {
		name := t.Name
		running := tunnelmgr.IsRunning(name)
		elapsed := ""
		if running {
			elapsed = tunnelElapsed(name)
		}

		label := tunnelLabel(name, running, elapsed)
		submenu := systray.NewMenu()

		actionLabel := locales.T("tray.start")
		if running {
			actionLabel = locales.T("tray.stop")
		}

		submenu.Add(actionLabel, func() {
			toggleTunnel(name, refresh)
		})
		submenu.Add(locales.T("tray.restart"), func() {
			restartTunnel(name, refresh)
		})
		submenu.Add(locales.T("tray.openLogs"), func() {
			openTunnelLog(name)
		})
		submenu.Add(locales.T("tray.edit"), func() {
			launchTUI(name)
		})

		menu.AddSubmenu(label, submenu)
	}

	menu.AddSeparator()
	menu.Add(locales.T("tray.openApp"), func() {
		launchTUI("")
	})
	menu.Add(locales.T("tray.testNotification"), func() {
		_ = notify.Notify(locales.T("app.title"), locales.T("tray.testNotificationBody"))
	})
	menu.AddSeparator()
	menu.Add(locales.T("tray.quit"), func() {
		active := activeTunnels(cfg)
		if len(active) > 0 {
			msg := fmt.Sprintf(locales.T("tray.stopBeforeQuit"), len(active))
			if askYesNo(locales.T("app.title"), msg) {
				for _, n := range active {
					_ = tunnelmgr.StopDaemon(n)
					_ = tunnelmgr.RemoveLogFile(n)
				}
			}
		}
		_ = tunnelmgr.RemoveTrayPidFile()
		tray.Remove()
		os.Exit(0)
	})

	return menu
}

func toggleTunnel(name string, refresh func()) {
	running := tunnelmgr.IsRunning(name)
	if running {
		if err := tunnelmgr.StopDaemon(name); err == nil {
			_ = tunnelmgr.RemoveLogFile(name)
			_ = notify.Notify(locales.T("app.title"), fmt.Sprintf(locales.T("messages.tunnelStopped"), name))
			refresh()
		} else {
			log.Printf("failed to stop tunnel %s: %v", name, err)
			_ = notify.Notify(locales.T("app.title"), fmt.Sprintf(locales.T("messages.failedToStop"), name, err))
		}
	} else {
		if err := startTunnelProcess(name); err != nil {
			_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("messages.failedToStart"), name, err))
		} else {
			_ = notify.Notify(locales.T("app.title"), fmt.Sprintf(locales.T("messages.tunnelStarted"), name))
			refresh()
		}
	}
}

func restartTunnel(name string, refresh func()) {
	if err := tunnelmgr.RestartTunnel(name, startTunnelProcess); err != nil {
		_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("messages.failedToRestart"), name, err))
	} else {
		_ = notify.Notify(locales.T("app.title"), fmt.Sprintf(locales.T("messages.tunnelRestarted"), name))
		refresh()
	}
}

func startTunnelProcess(name string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "tunnel-start", name)
	return cmd.Run()
}

func updateTooltip(tray *systray.SystemTray, cfg core.Config) {
	active := activeTunnels(cfg)
	count := len(active)

	if count == 0 {
		tray.SetTooltip(locales.T("tray.title"))
		return
	}
	if count == 1 {
		elapsed := tunnelElapsed(active[0])
		tray.SetTooltip(fmt.Sprintf(locales.T("tray.activeTooltip"), active[0], elapsed))
		return
	}
	tray.SetTooltip(fmt.Sprintf(locales.T("tray.activeCountTooltip"), count))
}

func tunnelLabel(name string, active bool, elapsed string) string {
	emoji := "⚪"
	if active {
		emoji = "🟢"
	}
	if active && elapsed != "" {
		return fmt.Sprintf("%s %s %s", emoji, name, elapsed)
	}
	return fmt.Sprintf("%s %s", emoji, name)
}

func activeTunnels(cfg core.Config) []string {
	var names []string
	seen := make(map[string]bool)
	for _, t := range cfg.Tunnels {
		if seen[t.Name] {
			continue
		}
		if tunnelmgr.IsRunning(t.Name) {
			names = append(names, t.Name)
			seen[t.Name] = true
		}
	}
	return names
}

func tunnelElapsed(name string) string {
	dir, err := tunnelmgr.PidDir()
	if err != nil {
		return "?"
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.pid", name))
	info, err := os.Stat(path)
	if err != nil {
		return "?"
	}
	elapsed := time.Since(info.ModTime())
	mins := int(elapsed.Minutes())
	secs := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func openTunnelLog(name string) {
	path, err := tunnelmgr.LogFilePath(name)
	if err != nil {
		_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedGetLogPath"), err))
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.noLogsYet"), name))
		return
	}
	if err := openFile(path); err != nil {
		_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedOpenLog"), err))
	}
}

// shellQuote returns a single-quoted shell string literal for s.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// launchTUI starts the wombat TUI in a new terminal process.
// If editTunnel is non-empty it opens directly to that tunnel's edit form.
func launchTUI(editTunnel string) {
	ctx := context.Background()
	exe, err := os.Executable()
	if err != nil {
		exe, err = exec.LookPath("wombat")
		if err != nil {
			log.Printf("launchTUI: could not locate wombat binary")
			_ = notify.Alert(locales.T("app.title"), locales.T("tray.binaryNotFound"))
			return
		}
	}
	args := []string{"tui"}
	if editTunnel != "" {
		args = append(args, "--edit-tunnel", editTunnel)
	}

	switch runtime.GOOS {
	case "darwin":
		quoted := []string{shellQuote(exe)}
		for _, a := range args {
			quoted = append(quoted, shellQuote(a))
		}
		cmdStr := strings.Join(quoted, " ")
		script := fmt.Sprintf(`tell application "Terminal" to do script %q`, cmdStr)
		if err := exec.CommandContext(ctx, "osascript", "-e", script).Start(); err != nil {
			log.Printf("launchTUI: failed to open Terminal: %v", err)
			_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedOpenTerminal"), err))
			return
		}
		_ = exec.CommandContext(ctx, "osascript", "-e", `tell application "Terminal" to activate`).Start()
	case "linux":
		terminals := []string{"x-terminal-emulator", "gnome-terminal", "kgx", "konsole", "xfce4-terminal", "xterm"}
		for _, term := range terminals {
			path, err := exec.LookPath(term)
			if err != nil {
				continue
			}
			var cmd *exec.Cmd
			switch term {
			case "gnome-terminal", "kgx":
				gtArgs := append([]string{"--", exe}, args...)
				cmd = exec.CommandContext(ctx, path, gtArgs...)
			default:
				termArgs := append([]string{"-e", exe}, args...)
				cmd = exec.CommandContext(ctx, path, termArgs...)
			}
			if err := cmd.Start(); err == nil {
				return
			}
		}
		// Fallback: spawn in background (may not have a visible terminal)
		cmd := exec.CommandContext(ctx, exe, args...)
		if err := cmd.Start(); err != nil {
			log.Printf("launchTUI: failed to start TUI: %v", err)
			_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedOpenTUI"), err))
		}
	case "windows":
		cmd := exec.CommandContext(ctx, "cmd", append([]string{"/c", "start", "", exe}, args...)...)
		if err := cmd.Start(); err != nil {
			log.Printf("launchTUI: failed to start TUI: %v", err)
			_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedOpenTUI"), err))
		}
	default:
		cmd := exec.CommandContext(ctx, exe, args...)
		if err := cmd.Start(); err != nil {
			log.Printf("launchTUI: failed to start TUI: %v", err)
			_ = notify.Alert(locales.T("app.title"), fmt.Sprintf(locales.T("tray.failedOpenTUI"), err))
		}
	}
}

// openFile opens the given file with the system's default application.
func openFile(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// askYesNo shows a native yes/no dialog. Returns true if the user chose Yes.
func askYesNo(title, message string) bool {
	switch runtime.GOOS {
	case "darwin":
		yes := locales.T("tray.yes")
		no := locales.T("tray.no")
		script := fmt.Sprintf(`display dialog %q buttons {%q, %q} default button %q with title %q`, message, yes, no, no, title)
		out, err := exec.Command("osascript", "-e", script).Output()
		return err == nil && strings.Contains(string(out), yes)
	case "linux":
		cmd := exec.Command("zenity", "--question", "--title", title, "--text", message)
		err := cmd.Run()
		return err == nil
	case "windows":
		script := fmt.Sprintf(
			`Add-Type -AssemblyName System.Windows.Forms; $r=[System.Windows.Forms.MessageBox]::Show(%q,%q,"YesNo"); exit [int]($r -ne "Yes")`,
			message, title,
		)
		err := exec.Command("powershell", "-Command", script).Run()
		return err == nil
	}
	return false
}
