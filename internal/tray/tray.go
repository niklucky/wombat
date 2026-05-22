package tray

import (
	"fmt"
	"os"

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
	systray.SetTitle("Wombat")
	systray.SetTooltip("Wombat SSH Helper")

	if data, err := os.ReadFile("assets/icon.png"); err == nil {
		systray.SetIcon(data)
	}

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

	go func() {
		for {
			select {
			case <-mNotify.ClickedCh:
				notify.Notify("Wombat", "Hello from the tray!")
			case <-mQuit.ClickedCh:
				systray.Quit()
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
				if mgr.IsActive(n) {
					if err := mgr.Stop(n); err == nil {
						tunnel.Active = false
						cfg.Save()
						notify.Notify("Wombat", fmt.Sprintf("Tunnel %s stopped", n))
					}
				} else {
					if err := mgr.Start(*tunnel, *host); err == nil {
						tunnel.Active = true
						cfg.Save()
						notify.Notify("Wombat", fmt.Sprintf("Tunnel %s started", n))
					}
				}
			}
		}(name, item)
	}
}

func onExit() {
	// Cleanup
}
