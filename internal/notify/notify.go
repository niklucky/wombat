//go:build !darwin

package notify

import (
	"github.com/gen2brain/beeep"
	"github.com/niklucky/wombat/assets"
)

// Notify sends a desktop notification.
func Notify(title, message string) error {
	return beeep.Notify(title, message, assets.NotificationIcon)
}

// Alert sends a desktop alert with sound.
func Alert(title, message string) error {
	return beeep.Alert(title, message, assets.NotificationIcon)
}
