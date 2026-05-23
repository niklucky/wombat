//go:build !gui

package main

import "fmt"

// runGUI reports that GUI support is not included in this build.
// It returns an error with the message "GUI not available: rebuild with -tags gui".
func runGUI() error {
	return fmt.Errorf("GUI not available: rebuild with -tags gui")
}
