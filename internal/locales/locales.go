// Package locales provides embedded language dictionaries and a simple
// key-based lookup helper for Wombat's user-facing strings.
package locales

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	//go:embed en.json
	enJSON []byte

	//go:embed ru.json
	ruJSON []byte

	//go:embed fr.json
	frJSON []byte

	//go:embed es.json
	esJSON []byte

	//go:embed zh.json
	zhJSON []byte
)

// Locale holds the loaded dictionary for a single language.
type Locale struct {
	strings map[string]string
}

var (
	mu        sync.RWMutex
	current   *Locale
	catalog   map[string]*Locale
	defaultLang = "en"
)

func init() {
	catalog = make(map[string]*Locale)

	for lang, data := range map[string][]byte{
		"en": enJSON,
		"ru": ruJSON,
		"fr": frJSON,
		"es": esJSON,
		"zh": zhJSON,
	} {
		var nested map[string]any
		if err := json.Unmarshal(data, &nested); err != nil {
			// Embedded JSON is static; panic on programmer error.
			panic(fmt.Sprintf("failed to load %s locale: %v", lang, err))
		}
		catalog[lang] = &Locale{strings: flatten("", nested)}
	}

	current = catalog[defaultLang]

	// Allow early localization of package-level strings (e.g. Cobra command
	// descriptions) via environment before the application config is loaded.
	if lang := os.Getenv("WOMBAT_LANG"); lang != "" {
		_ = SetLanguage(lang)
	}
}

// flatten recursively turns a nested JSON object into a flat
// "parent.child.leaf" key map of string values.
func flatten(prefix string, src map[string]any) map[string]string {
	out := make(map[string]string)
	for k, v := range src {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]any:
			for kk, vv := range flatten(key, val) {
				out[kk] = vv
			}
		}
	}
	return out
}

// SetLanguage switches the active locale. It returns an error if the
// requested language is not available.
func SetLanguage(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	l, ok := catalog[normalize(lang)]
	if !ok {
		return fmt.Errorf("unsupported language: %s", lang)
	}
	current = l
	return nil
}

// normalize turns a locale string like "ru_RU.UTF-8" into "ru".
func normalize(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if idx := strings.Index(lang, "."); idx != -1 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, "_"); idx != -1 {
		lang = lang[:idx]
	}
	return lang
}

// Current returns the active language code.
func Current() string {
	mu.RLock()
	defer mu.RUnlock()

	for code, l := range catalog {
		if l == current {
			return code
		}
	}
	return defaultLang
}

// Languages returns the list of supported language codes.
func Languages() []string {
	mu.RLock()
	defer mu.RUnlock()

	langs := make([]string, 0, len(catalog))
	for code := range catalog {
		langs = append(langs, code)
	}
	return langs
}

// Errorf looks up a key and returns it as an error. It is a convenience
// wrapper that avoids vet warnings about non-constant format strings.
func Errorf(key string, args ...any) error {
	return errors.New(T(key, args...))
}

// T looks up a key in the active locale and formats it with the supplied
// arguments. Missing keys fall back to English, then return the raw key.
func T(key string, args ...any) string {
	mu.RLock()
	l := current
	mu.RUnlock()

	return l.translate(key, args...)
}

func (l *Locale) translate(key string, args ...any) string {
	s, ok := l.strings[key]
	if !ok {
		s = catalog[defaultLang].strings[key]
	}
	if s == "" {
		s = key
	}
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}
	return s
}

// DetectLanguage returns the best language to use. Priority:
// 1. WOMBAT_LANG environment variable
// 2. LC_ALL / LANG environment variables
// 3. Default "en"
func DetectLanguage() string {
	for _, src := range []string{
		os.Getenv("WOMBAT_LANG"),
		os.Getenv("LC_ALL"),
		os.Getenv("LANG"),
	} {
		lang := normalize(src)
		if lang == "" {
			continue
		}
		if _, ok := catalog[lang]; ok {
			return lang
		}
	}
	return defaultLang
}
