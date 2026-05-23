//go:build gui

package main

import (
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/gui"
)

// runGUI loads the default application configuration and starts the GUI.
// It returns an error if configuration loading fails; otherwise it starts the GUI and returns nil.
func runGUI() error {
	cfg := core.DefaultConfig()
	if err := cfg.Load(); err != nil {
		return err
	}
	gui.Run(cfg)
	return nil
}
