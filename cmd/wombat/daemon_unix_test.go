//go:build darwin || linux

package main

import (
	"syscall"
	"testing"
)

func TestDaemonSysProcAttr_notNil(t *testing.T) {
	attr := daemonSysProcAttr()
	if attr == nil {
		t.Fatal("expected daemonSysProcAttr() to return non-nil *syscall.SysProcAttr")
	}
}

func TestDaemonSysProcAttr_setsidEnabled(t *testing.T) {
	attr := daemonSysProcAttr()
	if attr == nil {
		t.Fatal("expected daemonSysProcAttr() to return non-nil *syscall.SysProcAttr")
	}
	if !attr.Setsid {
		t.Error("expected SysProcAttr.Setsid to be true so daemon starts its own session")
	}
}

func TestDaemonSysProcAttr_returnsCorrectType(t *testing.T) {
	attr := daemonSysProcAttr()
	var _ *syscall.SysProcAttr = attr // compile-time type check
	if attr == nil {
		t.Fatal("expected non-nil return value")
	}
}