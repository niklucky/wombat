//go:build windows

package main

import "syscall"

// daemonSysProcAttr returns a *syscall.SysProcAttr configured to start a child process in a new process group on Windows.
func daemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
