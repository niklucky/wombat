//go:build !darwin

package notify

import (
	"github.com/gen2brain/beeep"
	"github.com/niklucky/wombat/assets"
)

// Notify sends a desktop notification with the given title and message using the package notification icon.
// It returns an error if the notification could not be delivered.
func Notify(title, message string) error {
	return beeep.Notify(title, message, assets.NotificationIcon)
}

// Alert sends a desktop alert with sound using the package notification icon.
// It returns an error if the alert could not be delivered.
func Alert(title, message string) error {
	return beeep.Alert(title, message, assets.NotificationIcon)
}
