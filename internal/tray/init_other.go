//go:build !darwin

package tray

// ensureNSApplicationInitialized is a no-op on non-Apple platforms.
func ensureNSApplicationInitialized() error {
	return nil
}
