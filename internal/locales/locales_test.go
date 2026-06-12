package locales

import (
	"os"
	"strings"
	"testing"
)

func setLanguage(t *testing.T, lang string) {
	t.Helper()
	if err := SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage(%q) failed: %v", lang, err)
	}
}

func unsetenv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) failed: %v", key, err)
	}
}

func setenv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Setenv(%q, %q) failed: %v", key, value, err)
	}
}

func TestLanguages(t *testing.T) {
	langs := Languages()
	if len(langs) != 5 {
		t.Fatalf("expected 5 languages, got %d: %v", len(langs), langs)
	}
	seen := make(map[string]bool)
	for _, l := range langs {
		seen[l] = true
	}
	for _, want := range []string{"en", "ru", "fr", "es", "zh"} {
		if !seen[want] {
			t.Fatalf("expected %s in languages, got %v", want, langs)
		}
	}
}

func TestSetLanguage(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	setLanguage(t, "ru")
	if Current() != "ru" {
		t.Fatalf("expected current language ru, got %s", Current())
	}

	setLanguage(t, "en")
	if Current() != "en" {
		t.Fatalf("expected current language en, got %s", Current())
	}
}

func TestSetLanguageUnsupported(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	if err := SetLanguage("zz"); err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

func TestTLookup(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	setLanguage(t, "en")
	if got := T("app.title"); got != "Wombat SSH Helper" {
		t.Fatalf("expected English title, got %q", got)
	}

	setLanguage(t, "ru")
	if got := T("app.title"); got != "Wombat SSH Helper" {
		t.Fatalf("expected Russian title to fall back to English, got %q", got)
	}
	if got := T("tabs.tunnels"); got != "Туннели" {
		t.Fatalf("expected Russian tunnels tab, got %q", got)
	}

	setLanguage(t, "fr")
	if got := T("tabs.tunnels"); got != "Tunnels" {
		t.Fatalf("expected French tunnels tab, got %q", got)
	}
	if got := T("actions.quit"); got != "Quitter" {
		t.Fatalf("expected French quit action, got %q", got)
	}

	setLanguage(t, "es")
	if got := T("tabs.tunnels"); got != "Túneles" {
		t.Fatalf("expected Spanish tunnels tab, got %q", got)
	}
	if got := T("actions.quit"); got != "Salir" {
		t.Fatalf("expected Spanish quit action, got %q", got)
	}

	setLanguage(t, "zh")
	if got := T("tabs.tunnels"); got != "隧道" {
		t.Fatalf("expected Chinese tunnels tab, got %q", got)
	}
	if got := T("actions.quit"); got != "退出" {
		t.Fatalf("expected Chinese quit action, got %q", got)
	}
}

func TestTArgs(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	setLanguage(t, "en")
	if got := T("messages.tunnelStarted", "web"); !strings.Contains(got, "web") {
		t.Fatalf("expected formatted message to contain name, got %q", got)
	}

	setLanguage(t, "ru")
	if got := T("messages.tunnelStarted", "web"); !strings.Contains(got, "web") {
		t.Fatalf("expected Russian formatted message to contain name, got %q", got)
	}
}

func TestTFallback(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	setLanguage(t, "ru")
	if got := T("this.key.does.not.exist"); got != "this.key.does.not.exist" {
		t.Fatalf("expected missing key to be returned as-is, got %q", got)
	}
}

func TestErrorf(t *testing.T) {
	defer func() { setLanguage(t, "en") }()

	setLanguage(t, "en")
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
		unsetenv(t, "WOMBAT_LANG")
		unsetenv(t, "LC_ALL")
		unsetenv(t, "LANG")
		setLanguage(t, "en")
	}()

	unsetenv(t, "WOMBAT_LANG")
	unsetenv(t, "LC_ALL")
	unsetenv(t, "LANG")
	if got := DetectLanguage(); got != "en" {
		t.Fatalf("expected default en, got %s", got)
	}

	setenv(t, "WOMBAT_LANG", "ru")
	if got := DetectLanguage(); got != "ru" {
		t.Fatalf("expected ru from WOMBAT_LANG, got %s", got)
	}

	unsetenv(t, "WOMBAT_LANG")
	setenv(t, "LANG", "ru_RU.UTF-8")
	if got := DetectLanguage(); got != "ru" {
		t.Fatalf("expected ru from LANG, got %s", got)
	}

	setenv(t, "LANG", "unknown")
	if got := DetectLanguage(); got != "en" {
		t.Fatalf("expected fallback en for unknown locale, got %s", got)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"en":          "en",
		"EN":          "en",
		"ru_RU":       "ru",
		"ru_RU.UTF-8": "ru",
		"  RU  ":      "ru",
		"es-MX":       "es",
		"zh-CN":       "zh",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Fatalf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
