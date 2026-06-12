package tunnelmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
)

// PidDir returns the directory where PID files are stored.
func PidDir() (string, error) {
	home, err := core.AppHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "pids"), nil
}

// LogDir returns the directory where tunnel logs are stored.
func LogDir() (string, error) {
	home, err := core.AppHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "logs"), nil
}

// pidFilePath returns the path to a tunnel's PID file.
func pidFilePath(name string) (string, error) {
	dir, err := PidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s.pid", name)), nil
}

// LogFilePath returns the path to a tunnel's log file.
func LogFilePath(name string) (string, error) {
	dir, err := LogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s.log", name)), nil
}

// WritePidFile writes the current process PID to the tunnel's PID file.
func WritePidFile(name string) error {
	path, err := pidFilePath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600)
}

// RemovePidFile removes a tunnel's PID file.
func RemovePidFile(name string) error {
	path, err := pidFilePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// ReadPid reads the PID from a tunnel's PID file.
func ReadPid(name string) (int, error) {
	path, err := pidFilePath(name)
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// IsRunning checks if a tunnel's daemon process is alive.
// It also cleans up stale PID files for dead processes.
func IsRunning(name string) bool {
	pid, err := ReadPid(name)
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = RemovePidFile(name)
		return false
	}
	// Signal 0 checks if process exists without sending a real signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		_ = RemovePidFile(name)
		return false
	}
	return true
}

// StopDaemon sends a termination signal to a tunnel's daemon process.
func StopDaemon(name string) error {
	pid, err := ReadPid(name)
	if err != nil {
		return fmt.Errorf("tunnel %q is not running", name)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = RemovePidFile(name)
		return fmt.Errorf("tunnel %q process not found", name)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		// Fallback to stronger signal
		_ = proc.Signal(os.Kill)
	}
	_ = RemovePidFile(name)
	return nil
}

// OpenLogFile opens the log file for a tunnel, creating directories if needed.
func OpenLogFile(name string) (*os.File, error) {
	path, err := LogFilePath(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
}

// RemoveLogFile deletes a tunnel's log file.
func RemoveLogFile(name string) error {
	path, err := LogFilePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// TrayPidFilePath returns the path to the tray daemon PID file.
func TrayPidFilePath() (string, error) {
	dir, err := PidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tray-daemon.pid"), nil
}

// WriteTrayPidFile writes the current process PID to the tray daemon PID file.
func WriteTrayPidFile() error {
	path, err := TrayPidFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600)
}

// RemoveTrayPidFile removes the tray daemon PID file.
func RemoveTrayPidFile() error {
	path, err := TrayPidFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// RestartTunnel stops a tunnel if it is running and then calls start.
func RestartTunnel(name string, start func(string) error) error {
	if IsRunning(name) {
		if err := StopDaemon(name); err != nil {
			return fmt.Errorf("stop tunnel %q: %w", name, err)
		}
	}
	if err := start(name); err != nil {
		return fmt.Errorf("start tunnel %q: %w", name, err)
	}
	return nil
}

// StopTrayDaemon sends a termination signal to the tray daemon process.
func StopTrayDaemon() error {
	path, err := TrayPidFilePath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return locales.Errorf("errors.trayDaemonNotRunning")
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		_ = os.Remove(path)
		return locales.Errorf("errors.invalidTrayPid")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(path)
		return locales.Errorf("errors.trayProcessNotFound")
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		_ = proc.Signal(os.Kill)
	} else {
		for i := 0; i < 50; i++ {
			time.Sleep(100 * time.Millisecond)
			if proc.Signal(syscall.Signal(0)) != nil {
				break
			}
		}
		if proc.Signal(syscall.Signal(0)) == nil {
			_ = proc.Signal(os.Kill)
			time.Sleep(200 * time.Millisecond)
		}
	}
	_ = os.Remove(path)
	return nil
}

// IsTrayRunning checks if the tray daemon process is alive.
// It also cleans up stale PID files for dead processes.
func IsTrayRunning() bool {
	path, err := TrayPidFilePath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		_ = os.Remove(path)
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(path)
		return false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		_ = os.Remove(path)
		return false
	}
	return true
}
