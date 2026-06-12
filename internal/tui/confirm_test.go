package tui

import (
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/locales"
)

func TestRenderConfirmDelete(t *testing.T) {
	if err := locales.SetLanguage("en"); err != nil {
		t.Fatalf("SetLanguage(en) failed: %v", err)
	}
	s := renderConfirmDelete("tunnel", "mytunnel")
	want := locales.T("dialog.deleteTitle", "tunnel", "mytunnel")
	if !strings.Contains(s, want) {
		t.Errorf("expected confirmation text, got: %q", s)
	}
	if !strings.Contains(s, "y") {
		t.Error("expected 'y' prompt")
	}
	if !strings.Contains(s, "n") {
		t.Error("expected 'n' prompt")
	}
}
