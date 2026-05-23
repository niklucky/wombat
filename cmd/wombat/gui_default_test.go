//go:build !gui

package main

import (
	"strings"
	"testing"
)

func TestRunGUI_returnsError(t *testing.T) {
	err := runGUI()
	if err == nil {
		t.Fatal("expected runGUI() to return an error when built without -tags gui, got nil")
	}
}

func TestRunGUI_errorContainsNotAvailable(t *testing.T) {
	err := runGUI()
	if err == nil {
		t.Fatal("expected runGUI() to return an error")
	}
	if !strings.Contains(err.Error(), "GUI not available") {
		t.Errorf("expected error to contain %q, got: %q", "GUI not available", err.Error())
	}
}

func TestRunGUI_errorContainsBuildInstructions(t *testing.T) {
	err := runGUI()
	if err == nil {
		t.Fatal("expected runGUI() to return an error")
	}
	const want = "GUI not available: rebuild with -tags gui"
	if err.Error() != want {
		t.Errorf("expected error %q, got: %q", want, err.Error())
	}
}