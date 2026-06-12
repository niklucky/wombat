package tui

import (
	"os"
	"testing"

	"github.com/niklucky/wombat/internal/locales"
)

func TestMain(m *testing.M) {
	if err := locales.SetLanguage("en"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestBoolToYesNo(t *testing.T) {
	if boolToYesNo(true) != "yes" {
		t.Error("expected 'yes' for true")
	}
	if boolToYesNo(false) != "no" {
		t.Error("expected 'no' for false")
	}
}

func TestYesNoToBool(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"yes", true},
		{"YES", true},
		{"y", true},
		{"true", true},
		{"no", false},
		{"NO", false},
		{"n", false},
		{"false", false},
		{"", false},
		{"maybe", false},
	}
	for _, c := range cases {
		got := yesNoToBool(c.input)
		if got != c.want {
			t.Errorf("yesNoToBool(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
