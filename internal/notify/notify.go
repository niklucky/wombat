package notify

import "github.com/gen2brain/beeep"

// Notify sends a desktop notification.
func Notify(title, message string) error {
	return beeep.Notify(title, message, "")
}

// Alert sends a desktop alert with sound.
func Alert(title, message string) error {
	return beeep.Alert(title, message, "")
}
