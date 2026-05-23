//go:build windows

package platform

import "syscall"

// DaemonSysProcAttr returns process attributes that detach a child process
// from the parent console session (Windows).
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
