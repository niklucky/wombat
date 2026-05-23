//go:build darwin || linux

package main

import "syscall"

// daemonSysProcAttr returns a *syscall.SysProcAttr configured to create a new
// session for a spawned process (Setsid = true), suitable for daemonizing on
// Darwin and Linux.
func daemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
