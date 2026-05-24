package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/systray"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

// tunnelMenu holds the menu items for a single tunnel.
type tunnelMenu struct {
	item      *systray.MenuItem
	startStop *systray.MenuItem
	restart   *systray.MenuItem
	openLogs  *systray.MenuItem
}

// RunWithTunnels starts the system tray application with tunnel support.
func RunWithTunnels(cfg core.Config) {
	systray.Run(func() { onReady(cfg) }, onExit)
}

func onReady(cfg core.Config) {
	_ = tunnelmgr.WriteTrayPidFile()

	if data, err := os.ReadFile("assets/tray-icon.png"); err == nil {
		systray.SetIcon(data)
	}

	updateTrayStatus(cfg)

	// Create a top-level menu item for each tunnel with a submenu
	tunnelMenus := make(map[string]*tunnelMenu)
	for _, t := range cfg.Tunnels {
		tm := &tunnelMenu{}
		tm.item = systray.AddMenuItem(
			tunnelLabel(t.Name, false, ""),
			fmt.Sprintf("Tunnel %s", t.Name),
		)
		tm.startStop = tm.item.AddSubMenuItem("Start", "Start or stop tunnel")
		tm.restart = tm.item.AddSubMenuItem("Restart", "Restart tunnel")
		tm.openLogs = tm.item.AddSubMenuItem("Open logs", "Open tunnel log file")
		tunnelMenus[t.Name] = tm
	}

	systray.AddSeparator()
	mNotify := systray.AddMenuItem("Test Notification", "Send a test notification")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Wombat")

	// Start ticker to update titles and tunnel menu labels
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			updateTrayStatus(cfg)
			updateTunnelMenuLabels(cfg, tunnelMenus)
		}
	}()

	// Initial label update so emojis reflect current state immediately
	updateTunnelMenuLabels(cfg, tunnelMenus)

	go func() {
		for {
			select {
			case <-mNotify.ClickedCh:
				notify.Notify("Wombat", "Hello from the tray!")
			case <-mQuit.ClickedCh:
				active := activeTunnels(cfg)
				if len(active) > 0 {
					msg := fmt.Sprintf("Stop %d active tunnel(s) before quitting?", len(active))
					if askYesNo("Wombat", msg) {
						for _, name := range active {
							_ = tunnelmgr.StopDaemon(name)
							_ = tunnelmgr.RemoveLogFile(name)
						}
					}
				}
				systray.Quit()
				return
			}
		}
	}()

	// Handle tunnel start/stop, restart, and open logs clicks
	for name, tm := range tunnelMenus {
		go func(n string, menu *tunnelMenu) {
			for range menu.startStop.ClickedCh {
				toggleTunnel(cfg, n, tunnelMenus)
			}
		}(name, tm)

		go func(n string, menu *tunnelMenu) {
			for range menu.restart.ClickedCh {
				restartTunnel(cfg, n, tunnelMenus)
			}
		}(name, tm)

		go func(n string, menu *tunnelMenu) {
			for range menu.openLogs.ClickedCh {
				openTunnelLog(n)
			}
		}(name, tm)
	}
}

func toggleTunnel(cfg core.Config, name string, menus map[string]*tunnelMenu) {
	running := tunnelmgr.IsRunning(name)
	if running {
		if err := tunnelmgr.StopDaemon(name); err == nil {
			_ = tunnelmgr.RemoveLogFile(name)
			notify.Alert("Wombat", fmt.Sprintf("Tunnel %s stopped", name))
			updateTrayStatus(cfg)
			updateTunnelMenuLabels(cfg, menus)
		}
	} else {
		if err := startTunnelProcess(name); err != nil {
			notify.Alert("Wombat", fmt.Sprintf("Failed to start tunnel %s: %v", name, err))
		} else {
			notify.Notify("Wombat", fmt.Sprintf("Tunnel %s started", name))
			updateTrayStatus(cfg)
			updateTunnelMenuLabels(cfg, menus)
		}
	}
}

func restartTunnel(cfg core.Config, name string, menus map[string]*tunnelMenu) {
	if err := tunnelmgr.RestartTunnel(name, startTunnelProcess); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("Failed to restart tunnel %s: %v", name, err))
	} else {
		notify.Notify("Wombat", fmt.Sprintf("Tunnel %s restarted", name))
		updateTrayStatus(cfg)
		updateTunnelMenuLabels(cfg, menus)
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

func onExit() {
	_ = tunnelmgr.RemoveTrayPidFile()
}

func updateTrayStatus(cfg core.Config) {
	active := activeTunnels(cfg)
	count := len(active)

	if count == 0 {
		systray.SetTitle("")
		systray.SetTooltip("Wombat SSH Helper")
	} else if count == 1 {
		elapsed := tunnelElapsed(active[0])
		systray.SetTitle(elapsed)
		systray.SetTooltip(fmt.Sprintf("%s active", active[0]))
	} else {
		systray.SetTitle(fmt.Sprintf("%d", count))
		systray.SetTooltip(fmt.Sprintf("%d tunnels active", count))
	}
}

func updateTunnelMenuLabels(cfg core.Config, menus map[string]*tunnelMenu) {
	for _, t := range cfg.Tunnels {
		menu, ok := menus[t.Name]
		if !ok {
			continue
		}
		running := tunnelmgr.IsRunning(t.Name)
		elapsed := ""
		if running {
			elapsed = tunnelElapsed(t.Name)
		}
		menu.item.SetTitle(tunnelLabel(t.Name, running, elapsed))
		if running {
			menu.startStop.SetTitle("Stop")
		} else {
			menu.startStop.SetTitle("Start")
		}
	}
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
		notify.Alert("Wombat", fmt.Sprintf("Failed to get log path: %v", err))
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		notify.Alert("Wombat", fmt.Sprintf("No logs for tunnel %s yet", name))
		return
	}
	if err := openFile(path); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("Failed to open log: %v", err))
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
		script := fmt.Sprintf(`display dialog %q buttons {"Yes", "No"} default button "No" with title %q`, message, title)
		out, err := exec.Command("osascript", "-e", script).Output()
		return err == nil && strings.Contains(string(out), "Yes")
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
