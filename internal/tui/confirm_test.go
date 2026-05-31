package tui

import (
	"strings"
	"testing"
)

func TestRenderConfirmDelete(t *testing.T) {
	s := renderConfirmDelete("tunnel", "mytunnel")
	if !strings.Contains(s, "Delete tunnel \"mytunnel\"?") {
		t.Errorf("expected confirmation text, got: %q", s)
	}
	if !strings.Contains(s, "y") {
		t.Error("expected 'y' prompt")
	}
	if !strings.Contains(s, "n") {
		t.Error("expected 'n' prompt")
	}
}
