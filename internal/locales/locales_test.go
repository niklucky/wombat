package locales

import (
	"os"
	"strings"
	"testing"
)

func TestLanguages(t *testing.T) {
	langs := Languages()
	if len(langs) != 4 {
		t.Fatalf("expected 4 languages, got %d: %v", len(langs), langs)
	}
	seen := make(map[string]bool)
	for _, l := range langs {
		seen[l] = true
	}
	for _, want := range []string{"en", "ru", "fr", "es"} {
		if !seen[want] {
			t.Fatalf("expected %s in languages, got %v", want, langs)
		}
	}
}

func TestSetLanguage(t *testing.T) {
	defer SetLanguage("en")

	if err := SetLanguage("ru"); err != nil {
		t.Fatalf("SetLanguage(ru) failed: %v", err)
	}
	if Current() != "ru" {
		t.Fatalf("expected current language ru, got %s", Current())
	}

	if err := SetLanguage("en"); err != nil {
		t.Fatalf("SetLanguage(en) failed: %v", err)
	}
	if Current() != "en" {
		t.Fatalf("expected current language en, got %s", Current())
	}
}

func TestSetLanguageUnsupported(t *testing.T) {
	defer SetLanguage("en")

	if err := SetLanguage("zz"); err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

func TestTLookup(t *testing.T) {
	defer SetLanguage("en")

	SetLanguage("en")
	if got := T("app.title"); got != "Wombat SSH Helper" {
		t.Fatalf("expected English title, got %q", got)
	}

	SetLanguage("ru")
	if got := T("app.title"); got != "Wombat SSH Helper" {
		t.Fatalf("expected Russian title to fall back to English, got %q", got)
	}
	if got := T("tabs.tunnels"); got != "Туннели" {
		t.Fatalf("expected Russian tunnels tab, got %q", got)
	}

	SetLanguage("fr")
	if got := T("tabs.tunnels"); got != "Tunnels" {
		t.Fatalf("expected French tunnels tab, got %q", got)
	}
	if got := T("actions.quit"); got != "Quitter" {
		t.Fatalf("expected French quit action, got %q", got)
	}

	SetLanguage("es")
	if got := T("tabs.tunnels"); got != "Túneles" {
		t.Fatalf("expected Spanish tunnels tab, got %q", got)
	}
	if got := T("actions.quit"); got != "Salir" {
		t.Fatalf("expected Spanish quit action, got %q", got)
	}
}

func TestTArgs(t *testing.T) {
	defer SetLanguage("en")

	SetLanguage("en")
	if got := T("messages.tunnelStarted", "web"); !strings.Contains(got, "web") {
		t.Fatalf("expected formatted message to contain name, got %q", got)
	}

	SetLanguage("ru")
	if got := T("messages.tunnelStarted", "web"); !strings.Contains(got, "web") {
		t.Fatalf("expected Russian formatted message to contain name, got %q", got)
	}
}

func TestTFallback(t *testing.T) {
	defer SetLanguage("en")

	SetLanguage("ru")
	if got := T("this.key.does.not.exist"); got != "this.key.does.not.exist" {
		t.Fatalf("expected missing key to be returned as-is, got %q", got)
	}
}

func TestErrorf(t *testing.T) {
	defer SetLanguage("en")

	SetLanguage("en")
	err := Errorf("errors.allFieldsRequired")
	if err == nil || err.Error() != "all fields are required" {
		t.Fatalf("unexpected error: %v", err)
	}

	err = Errorf("errors.invalidPort", "bad")
	if err == nil || !strings.Contains(err.Error(), "bad") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectLanguage(t *testing.T) {
	defer func() {
		os.Unsetenv("WOMBAT_LANG")
		os.Unsetenv("LC_ALL")
		os.Unsetenv("LANG")
		SetLanguage("en")
	}()

	os.Unsetenv("WOMBAT_LANG")
	os.Unsetenv("LC_ALL")
	os.Unsetenv("LANG")
	if got := DetectLanguage(); got != "en" {
		t.Fatalf("expected default en, got %s", got)
	}

	os.Setenv("WOMBAT_LANG", "ru")
	if got := DetectLanguage(); got != "ru" {
		t.Fatalf("expected ru from WOMBAT_LANG, got %s", got)
	}

	os.Unsetenv("WOMBAT_LANG")
	os.Setenv("LANG", "ru_RU.UTF-8")
	if got := DetectLanguage(); got != "ru" {
		t.Fatalf("expected ru from LANG, got %s", got)
	}

	os.Setenv("LANG", "unknown")
	if got := DetectLanguage(); got != "en" {
		t.Fatalf("expected fallback en for unknown locale, got %s", got)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"en":         "en",
		"EN":         "en",
		"ru_RU":      "ru",
		"ru_RU.UTF-8": "ru",
		"  RU  ":     "ru",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Fatalf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
