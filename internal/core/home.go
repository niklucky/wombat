package core

import (
	"os"
	"path/filepath"
	"strings"
)

// homePointerPath returns the fixed path to the app-home pointer file.
// This file always lives in the OS default config dir so we can find it
// regardless of where the user moves their app home.
func homePointerPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wombat", "home"), nil
}

// AppHome returns the configured app home directory.
// It reads from the pointer file, defaulting to the OS config dir.
func AppHome() (string, error) {
	path, err := homePointerPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			dir, _ := os.UserConfigDir()
			return filepath.Join(dir, "wombat"), nil
		}
		return "", err
	}
	home := strings.TrimSpace(string(data))
	if home == "" {
		dir, _ := os.UserConfigDir()
		return filepath.Join(dir, "wombat"), nil
	}
	return home, nil
}

// SetAppHome writes the app home pointer.
func SetAppHome(home string) error {
	path, err := homePointerPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(home), 0600)
}
