package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/sshconfig"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

// tunnelMenu holds the menu items for a single tunnel.
type tunnelMenu struct {
	item      *systray.MenuItem
	startStop *systray.MenuItem
	openLogs  *systray.MenuItem
}

// RunWithTunnels starts the system tray application with tunnel support.
func RunWithTunnels(cfg core.Config, mgr *tunnelmgr.Manager) {
	systray.Run(func() { onReady(cfg, mgr) }, onExit)
}

func onReady(cfg core.Config, mgr *tunnelmgr.Manager) {
	if data, err := os.ReadFile("assets/tray-icon.png"); err == nil {
		systray.SetIcon(data)
	}

	updateTrayStatus(cfg, mgr)

	// Create a top-level menu item for each tunnel with a submenu
	tunnelMenus := make(map[string]*tunnelMenu)
	for _, t := range cfg.Tunnels {
		tm := &tunnelMenu{}
		tm.item = systray.AddMenuItem(
			tunnelLabel(t.Name, false, ""),
			fmt.Sprintf("Tunnel %s", t.Name),
		)
		tm.startStop = tm.item.AddSubMenuItem("Start", "Start or stop tunnel")
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
			updateTrayStatus(cfg, mgr)
			updateTunnelMenuLabels(cfg, mgr, tunnelMenus)
		}
	}()

	// Initial label update so emojis reflect current state immediately
	updateTunnelMenuLabels(cfg, mgr, tunnelMenus)

	go func() {
		for {
			select {
			case <-mNotify.ClickedCh:
				notify.Notify("Wombat", "Hello from the tray!")
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()

	// Handle tunnel start/stop and open logs clicks
	for name, tm := range tunnelMenus {
		go func(n string, menu *tunnelMenu) {
			for range menu.startStop.ClickedCh {
				toggleTunnel(cfg, mgr, n, tunnelMenus)
			}
		}(name, tm)

		go func(n string, menu *tunnelMenu) {
			for range menu.openLogs.ClickedCh {
				openTunnelLog(n)
			}
		}(name, tm)
	}
}

func toggleTunnel(cfg core.Config, mgr *tunnelmgr.Manager, name string, menus map[string]*tunnelMenu) {
	tunnel := cfg.FindTunnel(name)
	if tunnel == nil {
		return
	}
	host := cfg.FindHost(tunnel.HostName)
	if host == nil {
		resolved, err := sshconfig.Resolve(tunnel.HostName)
		if err != nil {
			return
		}
		host = &resolved
	}

	running := tunnelmgr.IsRunning(name) || mgr.IsActive(name)
	if running {
		var stopped bool
		if tunnelmgr.IsRunning(name) {
			if err := tunnelmgr.StopDaemon(name); err == nil {
				stopped = true
			}
		}
		if mgr.IsActive(name) {
			if err := mgr.Stop(name); err == nil {
				stopped = true
			}
		}
		if stopped {
			notify.Notify("Wombat", fmt.Sprintf("Tunnel %s stopped", name))
			updateTrayStatus(cfg, mgr)
			updateTunnelMenuLabels(cfg, mgr, menus)
		}
	} else {
		if err := mgr.Start(*tunnel, *host); err != nil {
			notify.Alert("Wombat", fmt.Sprintf("Failed to start tunnel %s: %v", name, err))
		} else {
			notify.Notify("Wombat", fmt.Sprintf("Tunnel %s started", name))
			updateTrayStatus(cfg, mgr)
			updateTunnelMenuLabels(cfg, mgr, menus)
		}
	}
}

func onExit() {
	// Cleanup
}

func updateTrayStatus(cfg core.Config, mgr *tunnelmgr.Manager) {
	active := activeTunnels(cfg, mgr)
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

func updateTunnelMenuLabels(cfg core.Config, mgr *tunnelmgr.Manager, menus map[string]*tunnelMenu) {
	for _, t := range cfg.Tunnels {
		menu, ok := menus[t.Name]
		if !ok {
			continue
		}
		running := tunnelmgr.IsRunning(t.Name) || mgr.IsActive(t.Name)
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

func activeTunnels(cfg core.Config, mgr *tunnelmgr.Manager) []string {
	var names []string
	seen := make(map[string]bool)
	for _, t := range cfg.Tunnels {
		if seen[t.Name] {
			continue
		}
		if tunnelmgr.IsRunning(t.Name) || mgr.IsActive(t.Name) {
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
