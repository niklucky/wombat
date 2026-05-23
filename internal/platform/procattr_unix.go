//go:build darwin || linux

package platform

import "syscall"

// DaemonSysProcAttr returns process attributes that detach a child process
// from the parent terminal session (Unix).
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
