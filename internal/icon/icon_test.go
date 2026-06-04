package icon

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/niklucky/wombat/assets"
)

func TestGenerate(t *testing.T) {
	if assets.TrayIcon == nil {
		t.Skip("no tray icon asset available")
	}

	states := []State{
		{Connected: 0, Total: 0},
		{Connected: 0, Total: 3},
		{Connected: 1, Total: 3},
		{Connected: 3, Total: 3},
		{Connected: 2, Total: 2},
	}

	for _, s := range states {
		b, err := Generate(assets.TrayIcon, s)
		if err != nil {
			t.Fatalf("Generate(%+v) error: %v", s, err)
		}
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Generate(%+v) produced invalid PNG: %v", s, err)
		}
		bounds := img.Bounds()
		if bounds.Dx() == 0 || bounds.Dy() == 0 {
			t.Fatalf("Generate(%+v) produced empty image", s)
		}
	}
}

func TestBadgeColor(t *testing.T) {
	tests := []struct {
		state   State
		wantNil bool
	}{
		{State{0, 0}, true},
		{State{0, 3}, false},
		{State{1, 3}, false},
		{State{3, 3}, false},
	}

	for _, tt := range tests {
		_, _, _, a := tt.state.BadgeColor().RGBA()
		if tt.wantNil && a != 0 {
			t.Errorf("BadgeColor(%+v) expected transparent, got alpha %d", tt.state, a)
		}
		if !tt.wantNil && a == 0 {
			t.Errorf("BadgeColor(%+v) expected opaque, got transparent", tt.state)
		}
	}
}
