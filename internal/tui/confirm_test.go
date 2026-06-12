package tui

import (
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/locales"
)

func TestRenderConfirmDelete(t *testing.T) {
	_ = locales.SetLanguage("en")
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
