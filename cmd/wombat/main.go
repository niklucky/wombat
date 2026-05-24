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
	"github.com/niklucky/wombat/internal/sshconfig"
	"github.com/niklucky/wombat/internal/sshutil"
	"github.com/niklucky/wombat/internal/tray"
	"github.com/niklucky/wombat/internal/tui"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

var rootCmd = &cobra.Command{
	Use:   "wombat",
	Short: "Wombat — a cross-platform SSH helper",
	Long:  `Wombat is a cross-platform SSH helper with TUI, system tray, and optional GUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: start TUI + tray
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}

		// Start tray daemon in background process (only if enabled and not already running)
		if cfg.OpenTray && !tunnelmgr.IsTrayRunning() {
			exe, err := os.Executable()
			if err != nil {
				log.Printf("failed to locate executable for tray-daemon: %v", err)
			} else {
				trayDaemon := exec.Command(exe, "tray-daemon")
				trayDaemon.SysProcAttr = daemonSysProcAttr()
				if startErr := trayDaemon.Start(); startErr != nil {
					log.Printf("failed to start tray-daemon: %v", startErr)
				}
			}
		}

		// Start TUI
		p := tea.NewProgram(tui.NewModel(cfg), tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if m, ok := finalModel.(tui.Model); ok {
			if m.RestartRequired {
				fmt.Println("App home changed. Please restart Wombat.")
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
	Short: "List configured hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
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
	Short: "Add a new host",
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
			return fmt.Errorf("--address and --user are required")
		}

		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
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
		fmt.Printf("Host %q added.\n", args[0])
		return nil
	},
}

var removeHostCmd = &cobra.Command{
	Use:   "remove-host <name>",
	Short: "Remove a host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		cfg.RemoveHost(args[0])
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Host %q removed.\n", args[0])
		return nil
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: "Connect to a host via SSH",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		host := cfg.FindHost(args[0])
		if host == nil {
			resolved, err := sshconfig.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("host %q not found", args[0])
			}
			host = &resolved
		}
		return execSSH(*host)
	},
}

var testCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test TCP connectivity to a host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		host := cfg.FindHost(args[0])
		if host == nil {
			resolved, err := sshconfig.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("host %q not found", args[0])
			}
			host = &resolved
		}
		if err := sshutil.TestConnection(*host); err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			return err
		}
		fmt.Println("Connection OK")
		return nil
	},
}

var importSSHConfigCmd = &cobra.Command{
	Use:   "import-ssh-config",
	Short: "Import hosts from ~/.ssh/config into Wombat",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sshconfig.EnsureSetup(); err != nil {
			return err
		}
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}

		imported, err := sshconfig.ImportFromMainConfig()
		if err != nil {
			return err
		}
		if len(imported) == 0 {
			fmt.Println("No importable hosts found in ~/.ssh/config.")
			return nil
		}

		existing := make(map[string]bool)
		for _, h := range cfg.Hosts {
			existing[h.Name] = true
		}

		var added int
		for _, h := range imported {
			if existing[h.Name] {
				fmt.Printf("Skipping %q (already exists)\n", h.Name)
				continue
			}
			cfg.AddHost(h)
			fmt.Printf("Imported %q (%s@%s)\n", h.Name, h.User, h.Address)
			added++
		}

		if added > 0 {
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Imported %d host(s).\n", added)
		} else {
			fmt.Println("No new hosts to import.")
		}
		return nil
	},
}

var addTunnelCmd = &cobra.Command{
	Use:   "add-tunnel <name>",
	Short: "Add a new tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostName, _ := cmd.Flags().GetString("host")
		localPort, _ := cmd.Flags().GetInt("local-port")
		remoteHost, _ := cmd.Flags().GetString("remote-host")
		remotePort, _ := cmd.Flags().GetInt("remote-port")

		if hostName == "" || localPort == 0 || remotePort == 0 {
			return fmt.Errorf("--host, --local-port, and --remote-port are required")
		}
		if remoteHost == "" {
			remoteHost = "localhost"
		}

		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
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
		fmt.Printf("Tunnel %q added: localhost:%d -> %s:%d via %s\n",
			args[0], localPort, remoteHost, remotePort, hostName)
		return nil
	},
}

var removeTunnelCmd = &cobra.Command{
	Use:   "remove-tunnel <name>",
	Short: "Remove a tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		if tunnelmgr.IsRunning(args[0]) {
			return fmt.Errorf("tunnel %q is still running; stop it before removing", args[0])
		}
		cfg.RemoveTunnel(args[0])
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Tunnel %q removed.\n", args[0])
		return nil
	},
}

var tunnelDaemonCmd = &cobra.Command{
	Use:    "tunnel-daemon <name>",
	Short:  "Run a tunnel in the background (internal use)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		tunnel := cfg.FindTunnel(args[0])
		if tunnel == nil {
			return fmt.Errorf("tunnel %q not found", args[0])
		}
		host := cfg.FindHost(tunnel.HostName)
		if host == nil {
			resolved, err := sshconfig.Resolve(tunnel.HostName)
			if err != nil {
				return fmt.Errorf("host %q for tunnel not found", tunnel.HostName)
			}
			host = &resolved
		}

		// Open log file
		logFile, err := tunnelmgr.OpenLogFile(tunnel.Name)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)

		// Write PID file
		if err := tunnelmgr.WritePidFile(tunnel.Name); err != nil {
			log.Printf("ERROR: write PID file: %v", err)
			return fmt.Errorf("write PID file: %w", err)
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
	Short:  "Run the system tray daemon (internal use)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		tray.RunWithTunnels(cfg)
		return nil
	},
}

var tunnelStartCmd = &cobra.Command{
	Use:   "tunnel-start <name>",
	Short: "Start a tunnel in the background",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if tunnelmgr.IsRunning(name) {
			return fmt.Errorf("tunnel %q is already running", name)
		}

		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find executable: %w", err)
		}

		logFile, err := tunnelmgr.OpenLogFile(name)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		defer logFile.Close()

		daemonCmd := exec.Command(exe, "tunnel-daemon", name)
		daemonCmd.Stdout = logFile
		daemonCmd.Stderr = logFile
		daemonCmd.SysProcAttr = daemonSysProcAttr()

		if err := daemonCmd.Start(); err != nil {
			return fmt.Errorf("start daemon: %w", err)
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
				return fmt.Errorf("tunnel %q failed to start; log output:\n%s", name, string(logData))
			}
			return fmt.Errorf("tunnel %q failed to start (see log: %s)", name, logFile.Name())
		}

		fmt.Printf("Tunnel %q started in background (PID %d).\n", name, daemonCmd.Process.Pid)
		fmt.Printf("Logs: %s\n", logFile.Name())
		return nil
	},
}

var tunnelStopCmd = &cobra.Command{
	Use:   "tunnel-stop <name>",
	Short: "Stop a running tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !tunnelmgr.IsRunning(name) {
			_ = tunnelmgr.RemovePidFile(name)
			return fmt.Errorf("tunnel %q is not running", name)
		}
		if err := tunnelmgr.StopDaemon(name); err != nil {
			return err
		}
		fmt.Printf("Tunnel %q stopped.\n", name)
		return nil
	},
}

var tunnelRestartCmd = &cobra.Command{
	Use:   "tunnel-restart <name>",
	Short: "Restart a tunnel",
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

var tunnelListCmd = &cobra.Command{
	Use:   "tunnel-list",
	Short: "List all tunnels",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.DefaultConfig()
		if err := cfg.Load(); err != nil {
			return err
		}
		for _, t := range cfg.Tunnels {
			status := "inactive"
			if tunnelmgr.IsRunning(t.Name) {
				status = "active"
			}
			fmt.Printf("%s: %s [%s] localhost:%d -> %s:%d\n",
				t.Name, t.HostName, status, t.LocalPort, t.RemoteHost, t.RemotePort)
		}
		return nil
	},
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the GUI (requires -tags gui build)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGUI()
	},
}

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wombat %s\n", version)
	},
}

// init registers CLI subcommands on rootCmd and configures flags for host and tunnel creation.
// It attaches commands for host management, tunnel management, GUI/version display, and background daemons,
// and marks required flags for add-host (`address`, `user`) and add-tunnel (`host`, `local-port`, `remote-port`).
func init() {
	rootCmd.AddCommand(
		listCmd,
		addHostCmd, removeHostCmd,
		connectCmd, testCmd,
		importSSHConfigCmd,
		addTunnelCmd, removeTunnelCmd,
		tunnelStartCmd, tunnelStopCmd, tunnelRestartCmd, tunnelListCmd,
		guiCmd, versionCmd,
		tunnelDaemonCmd, trayDaemonCmd,
	)

	addHostCmd.Flags().String("address", "", "Host address (required)")
	addHostCmd.Flags().String("user", "", "SSH user (required)")
	addHostCmd.Flags().Int("port", 22, "SSH port")
	addHostCmd.Flags().String("key", "", "Path to private key")
	_ = addHostCmd.MarkFlagRequired("address")
	_ = addHostCmd.MarkFlagRequired("user")

	addTunnelCmd.Flags().String("host", "", "SSH host alias to tunnel through (required)")
	addTunnelCmd.Flags().Int("local-port", 0, "Local port to listen on (required)")
	addTunnelCmd.Flags().String("remote-host", "localhost", "Remote host to forward to")
	addTunnelCmd.Flags().Int("remote-port", 0, "Remote port to forward to (required)")
	_ = addTunnelCmd.MarkFlagRequired("host")
	_ = addTunnelCmd.MarkFlagRequired("local-port")
	_ = addTunnelCmd.MarkFlagRequired("remote-port")
}

// main executes the root Cobra command for the wombat CLI.
// If execution returns an error, main terminates the process with exit code 1.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
