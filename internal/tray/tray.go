package tray

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/systray"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/sshconfig"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

// RunWithTunnels starts the system tray application with tunnel support.
func RunWithTunnels(cfg core.Config, mgr *tunnelmgr.Manager) {
	systray.Run(func() { onReady(cfg, mgr) }, onExit)
}

func onReady(cfg core.Config, mgr *tunnelmgr.Manager) {
	if data, err := os.ReadFile("assets/tray-icon.png"); err == nil {
		systray.SetIcon(data)
	}

	updateTrayStatus(cfg, mgr)

	mTunnels := systray.AddMenuItem("Tunnels", "Manage tunnels")
	systray.AddSeparator()
	mNotify := systray.AddMenuItem("Test Notification", "Send a test notification")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Wombat")

	// Build tunnel menu items
	tunnelItems := make(map[string]*systray.MenuItem)
	for _, t := range cfg.Tunnels {
		label := fmt.Sprintf("%s (%d -> %s:%d)", t.Name, t.LocalPort, t.RemoteHost, t.RemotePort)
		item := mTunnels.AddSubMenuItem(label, "Toggle tunnel")
		tunnelItems[t.Name] = item
	}

	// Start ticker to update title
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			updateTrayStatus(cfg, mgr)
		}
	}()

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

	// Handle tunnel toggles
	for name, item := range tunnelItems {
		go func(n string, it *systray.MenuItem) {
			for range it.ClickedCh {
				tunnel := cfg.FindTunnel(n)
				if tunnel == nil {
					continue
				}
				host := cfg.FindHost(tunnel.HostName)
				if host == nil {
					resolved, err := sshconfig.Resolve(tunnel.HostName)
					if err != nil {
						continue
					}
					host = &resolved
				}

				running := tunnelmgr.IsRunning(n) || mgr.IsActive(n)
				if running {
					var stopped bool
					if tunnelmgr.IsRunning(n) {
						if err := tunnelmgr.StopDaemon(n); err == nil {
							stopped = true
						}
					}
					if mgr.IsActive(n) {
						if err := mgr.Stop(n); err == nil {
							stopped = true
						}
					}
					if stopped {
						notify.Notify("Wombat", fmt.Sprintf("Tunnel %s stopped", n))
						updateTrayStatus(cfg, mgr)
					}
				} else {
					if err := mgr.Start(*tunnel, *host); err != nil {
						notify.Alert("Wombat", fmt.Sprintf("Failed to start tunnel %s: %v", n, err))
					} else {
						notify.Notify("Wombat", fmt.Sprintf("Tunnel %s started", n))
						updateTrayStatus(cfg, mgr)
					}
				}
			}
		}(name, item)
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
