package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
	"github.com/niklucky/wombat/internal/sshconfig"
	"github.com/niklucky/wombat/internal/sshutil"
	"github.com/niklucky/wombat/internal/tray"
	"github.com/niklucky/wombat/internal/tui"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

var rootCmd = &cobra.Command{
	Use:   "wombat",
	Short: locales.T("app.shortDescription"),
	Long:  locales.T("app.longDescription"),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: start TUI + tray
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)

		// Restart tray daemon in background process (only if enabled)
		if cfg.OpenTray {
			if tunnelmgr.IsTrayRunning() {
				_ = tunnelmgr.StopTrayDaemon()
			}
			startTrayDaemon()
		}

		// Start TUI
		p := tea.NewProgram(tui.NewModel(cfg), tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if m, ok := finalModel.(tui.Model); ok {
			if m.RestartRequired {
				fmt.Println(locales.T("messages.appHomeChanged"))
				return nil
			}
			if m.SelectedHost != nil {
				return execSSH(*m.SelectedHost)
			}
		}
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: locales.T("cli.listHosts"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		for _, h := range cfg.Hosts {
			port := h.Port
			if port == 0 {
				port = 22
			}
			fmt.Printf("%s: %s@%s:%d\n", h.Name, h.User, h.Address, port)
		}
		return nil
	},
}

var addHostCmd = &cobra.Command{
	Use:   "add-host <name>",
	Short: locales.T("cli.addHost"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		address, _ := cmd.Flags().GetString("address")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		key, _ := cmd.Flags().GetString("key")

		if address == "" || user == "" {
			return locales.Errorf("cli.addressAndUserRequired")
		}

		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		cfg.AddHost(core.Host{
			Name:    args[0],
			Address: address,
			User:    user,
			Port:    port,
			KeyPath: key,
		})
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf(locales.T("cli.hostAdded")+"\n", args[0])
		return nil
	},
}

var removeHostCmd = &cobra.Command{
	Use:   "remove-host <name>",
	Short: locales.T("cli.removeHost"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		cfg.RemoveHost(args[0])
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf(locales.T("cli.hostRemoved")+"\n", args[0])
		return nil
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: locales.T("cli.connect"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		host := cfg.FindHost(args[0])
		if host == nil {
			resolved, err := sshconfig.Resolve(args[0])
			if err != nil {
				return locales.Errorf("cli.hostNotFound", args[0])
			}
			host = &resolved
		}
		return execSSH(*host)
	},
}

var testCmd = &cobra.Command{
	Use:   "test <name>",
	Short: locales.T("cli.test"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		host := cfg.FindHost(args[0])
		if host == nil {
			resolved, err := sshconfig.Resolve(args[0])
			if err != nil {
				return locales.Errorf("cli.hostNotFound", args[0])
			}
			host = &resolved
		}
		if err := sshutil.TestConnection(*host); err != nil {
			fmt.Printf(locales.T("cli.connectionFailed")+"\n", err)
			return err
		}
		fmt.Println(locales.T("cli.connectionOK"))
		return nil
	},
}

var importSSHConfigCmd = &cobra.Command{
	Use:   "import-ssh-config",
	Short: locales.T("cli.importSSHConfig"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)

		imported, err := sshconfig.ImportFromMainConfig()
		if err != nil {
			return err
		}
		if len(imported) == 0 {
			fmt.Println(locales.T("cli.noImportableHosts"))
			return nil
		}

		existing := make(map[string]bool)
		for _, h := range cfg.Hosts {
			existing[h.Name] = true
		}

		var added int
		for _, h := range imported {
			if existing[h.Name] {
				fmt.Printf(locales.T("cli.skippingExisting")+"\n", h.Name)
				continue
			}
			cfg.AddHost(h)
			fmt.Printf(locales.T("cli.importedHost")+"\n", h.Name, h.User, h.Address)
			added++
		}

		if added > 0 {
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf(locales.T("cli.importedCount")+"\n", added)
		} else {
			fmt.Println(locales.T("cli.noNewHosts"))
		}
		return nil
	},
}

var addTunnelCmd = &cobra.Command{
	Use:   "add-tunnel <name>",
	Short: locales.T("cli.addTunnel"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostName, _ := cmd.Flags().GetString("host")
		localPort, _ := cmd.Flags().GetInt("local-port")
		remoteHost, _ := cmd.Flags().GetString("remote-host")
		remotePort, _ := cmd.Flags().GetInt("remote-port")

		if hostName == "" || localPort == 0 || remotePort == 0 {
			return locales.Errorf("cli.hostLocalRemoteRequired")
		}
		if remoteHost == "" {
			remoteHost = "localhost"
		}

		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		cfg.AddTunnel(core.Tunnel{
			Name:       args[0],
			HostName:   hostName,
			LocalPort:  localPort,
			RemoteHost: remoteHost,
			RemotePort: remotePort,
		})
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf(locales.T("cli.tunnelAdded")+"\n",
			args[0], localPort, remoteHost, remotePort, hostName)
		return nil
	},
}

var removeTunnelCmd = &cobra.Command{
	Use:   "remove-tunnel <name>",
	Short: locales.T("cli.removeTunnel"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		if tunnelmgr.IsRunning(args[0]) {
			return locales.Errorf("cli.tunnelRunning", args[0])
		}
		cfg.RemoveTunnel(args[0])
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf(locales.T("cli.tunnelRemoved")+"\n", args[0])
		return nil
	},
}

var tunnelDaemonCmd = &cobra.Command{
	Use:    "tunnel-daemon <name>",
	Short:  locales.T("cli.tunnelDaemon"),
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		tunnel := cfg.FindTunnel(args[0])
		if tunnel == nil {
			return locales.Errorf("cli.tunnelNotFound", args[0])
		}
		host := cfg.FindHost(tunnel.HostName)
		if host == nil {
			resolved, err := sshconfig.Resolve(tunnel.HostName)
			if err != nil {
				return locales.Errorf("cli.hostForTunnelNotFound", tunnel.HostName)
			}
			host = &resolved
		}

		// Open log file
		logFile, err := tunnelmgr.OpenLogFile(tunnel.Name)
		if err != nil {
			return locales.Errorf("errors.openLogFile", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)

		// Write PID file
		if err := tunnelmgr.WritePidFile(tunnel.Name); err != nil {
			log.Printf("ERROR: write PID file: %v", err)
			return locales.Errorf("errors.writePidFile", err)
		}
		defer tunnelmgr.RemovePidFile(tunnel.Name)

		port := host.Port
		if port == 0 {
			port = 22
		}
		log.Printf("Starting tunnel %q: localhost:%d -> %s:%d via %s@%s:%d",
			tunnel.Name, tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort,
			host.User, host.Address, port)
		_ = logFile.Sync()

		mgr := tunnelmgr.NewManager()

		// Start tunnel in a goroutine so signals can be handled in main goroutine
		startErr := make(chan error, 1)
		go func() {
			if err := mgr.Start(*tunnel, *host); err != nil {
				startErr <- err
				return
			}
			startErr <- nil
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// Wait for tunnel to start or early signal
		select {
		case err := <-startErr:
			if err != nil {
				log.Printf("ERROR: failed to start tunnel: %v", err)
				_ = logFile.Sync()
				return err
			}
		case sig := <-sigCh:
			log.Printf("Received signal %v before tunnel started, exiting", sig)
			_ = logFile.Sync()
			return nil
		}

		log.Printf("Tunnel %q is active and forwarding connections", tunnel.Name)
		_ = logFile.Sync()

		// Wait for stop signal
		sig := <-sigCh
		log.Printf("Received signal %v, stopping tunnel %q", sig, tunnel.Name)
		_ = logFile.Sync()

		if err := mgr.Stop(tunnel.Name); err != nil {
			log.Printf("ERROR: failed to stop tunnel: %v", err)
			_ = logFile.Sync()
		}
		log.Printf("Tunnel %q stopped", tunnel.Name)
		_ = logFile.Sync()
		return nil
	},
}

var trayDaemonCmd = &cobra.Command{
	Use:    "tray-daemon",
	Short:  locales.T("cli.trayDaemon"),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		tray.RunWithTunnels(cfg)
		return nil
	},
}

var tunnelStartCmd = &cobra.Command{
	Use:   "tunnel-start <name>",
	Short: locales.T("cli.tunnelStart"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if tunnelmgr.IsRunning(name) {
			return locales.Errorf("cli.tunnelAlreadyRunning", name)
		}

		exe, err := os.Executable()
		if err != nil {
			return locales.Errorf("cli.findExecutable", err)
		}

		logFile, err := tunnelmgr.OpenLogFile(name)
		if err != nil {
			return locales.Errorf("errors.openLogFile", err)
		}
		defer logFile.Close()

		daemonCmd := exec.Command(exe, "tunnel-daemon", name)
		daemonCmd.Stdout = logFile
		daemonCmd.Stderr = logFile
		daemonCmd.SysProcAttr = daemonSysProcAttr()

		if err := daemonCmd.Start(); err != nil {
			return locales.Errorf("cli.startDaemon", err)
		}

		waitCh := make(chan error, 1)
		go func() {
			waitCh <- daemonCmd.Wait()
		}()

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(5 * time.Second)
		ready := false
	loop:
		for {
			select {
			case <-timeout:
				break loop
			case <-waitCh:
				break loop
			case <-ticker.C:
				if tunnelmgr.IsRunning(name) {
					ready = true
					break loop
				}
			}
		}

		if !ready {
			_ = logFile.Sync()
			logData, _ := os.ReadFile(logFile.Name())
			if len(logData) > 0 {
				return locales.Errorf("cli.tunnelFailedToStartWithLog", name, string(logData))
			}
			return locales.Errorf("cli.tunnelFailedToStart", name, logFile.Name())
		}

		fmt.Printf(locales.T("cli.tunnelStarted")+"\n", name, daemonCmd.Process.Pid)
		fmt.Printf(locales.T("cli.logs")+"\n", logFile.Name())
		return nil
	},
}

var tunnelStopCmd = &cobra.Command{
	Use:   "tunnel-stop <name>",
	Short: locales.T("cli.tunnelStop"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !tunnelmgr.IsRunning(name) {
			_ = tunnelmgr.RemovePidFile(name)
			return locales.Errorf("cli.tunnelNotRunning", name)
		}
		if err := tunnelmgr.StopDaemon(name); err != nil {
			return err
		}
		fmt.Printf(locales.T("cli.tunnelStopped")+"\n", name)
		return nil
	},
}

var tunnelRestartCmd = &cobra.Command{
	Use:   "tunnel-restart <name>",
	Short: locales.T("cli.tunnelRestart"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if tunnelmgr.IsRunning(name) {
			if err := tunnelmgr.StopDaemon(name); err != nil {
				return err
			}
		}
		return tunnelStartCmd.RunE(cmd, args)
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: locales.T("cli.tui"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)

		// Restart tray daemon when opening TUI
		if cfg.OpenTray {
			if tunnelmgr.IsTrayRunning() {
				_ = tunnelmgr.StopTrayDaemon()
			}
			startTrayDaemon()
		}

		editTunnelName, _ := cmd.Flags().GetString("edit-tunnel")
		p := tea.NewProgram(tui.NewModelWithEdit(cfg, editTunnelName), tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if m, ok := finalModel.(tui.Model); ok {
			if m.RestartRequired {
				fmt.Println(locales.T("messages.appHomeChanged"))
				return nil
			}
			if m.SelectedHost != nil {
				return execSSH(*m.SelectedHost)
			}
		}
		return nil
	},
}

var tunnelListCmd = &cobra.Command{
	Use:   "tunnel-list",
	Short: locales.T("cli.tunnelList"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		applyConfigLanguage(cfg)
		for _, t := range cfg.Tunnels {
			status := locales.T("status.inactive")
			if tunnelmgr.IsRunning(t.Name) {
				status = locales.T("status.active")
			}
			fmt.Printf("%s: %s [%s] localhost:%d -> %s:%d\n",
				t.Name, t.HostName, status, t.LocalPort, t.RemoteHost, t.RemotePort)
		}
		return nil
	},
}

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: locales.T("cli.version"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(locales.T("cli.versionFormat")+"\n", version)
	},
}

// applyConfigLanguage sets the active locale from config or system detection.
func applyConfigLanguage(cfg core.Config) {
	lang := cfg.Language
	if lang == "" {
		lang = locales.DetectLanguage()
	}
	_ = locales.SetLanguage(lang)
}

// startTrayDaemon launches the tray-daemon subcommand in the background.
func startTrayDaemon() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("failed to locate executable for tray-daemon: %v", err)
		return
	}
	trayDaemon := exec.Command(exe, "tray-daemon")
	trayDaemon.SysProcAttr = daemonSysProcAttr()
	if err := trayDaemon.Start(); err != nil {
		log.Printf("failed to start tray-daemon: %v", err)
	}
}

// init registers CLI subcommands on rootCmd and configures flags for host and tunnel creation.
// It attaches commands for host management, tunnel management, version display, and background daemons,
// and marks required flags for add-host (`address`, `user`) and add-tunnel (`host`, `local-port`, `remote-port`).
func init() {
	rootCmd.AddCommand(
		listCmd,
		addHostCmd, removeHostCmd,
		connectCmd, testCmd,
		importSSHConfigCmd,
		addTunnelCmd, removeTunnelCmd,
		tunnelStartCmd, tunnelStopCmd, tunnelRestartCmd, tunnelListCmd,
		versionCmd,
		tunnelDaemonCmd, trayDaemonCmd,
		tuiCmd,
	)

	addHostCmd.Flags().String("address", "", locales.T("flags.address"))
	addHostCmd.Flags().String("user", "", locales.T("flags.user"))
	addHostCmd.Flags().Int("port", 22, locales.T("flags.port"))
	addHostCmd.Flags().String("key", "", locales.T("flags.key"))
	_ = addHostCmd.MarkFlagRequired("address")
	_ = addHostCmd.MarkFlagRequired("user")

	addTunnelCmd.Flags().String("host", "", locales.T("flags.host"))
	addTunnelCmd.Flags().Int("local-port", 0, locales.T("flags.localPort"))
	addTunnelCmd.Flags().String("remote-host", "localhost", locales.T("flags.remoteHost"))
	addTunnelCmd.Flags().Int("remote-port", 0, locales.T("flags.remotePort"))
	_ = addTunnelCmd.MarkFlagRequired("host")
	_ = addTunnelCmd.MarkFlagRequired("local-port")
	_ = addTunnelCmd.MarkFlagRequired("remote-port")

	tuiCmd.Flags().String("edit-tunnel", "", locales.T("flags.editTunnel"))
}

// main executes the root Cobra command for the wombat CLI.
// If execution returns an error, main terminates the process with exit code 1.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
