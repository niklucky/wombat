//go:build !gui

package main

import "fmt"

func runGUI() error {
	return fmt.Errorf("GUI not available: rebuild with -tags gui")
}
