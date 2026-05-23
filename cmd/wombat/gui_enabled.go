//go:build gui

package main

import (
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/gui"
)

func runGUI() error {
	cfg := core.DefaultConfig()
	if err := cfg.Load(); err != nil {
		return err
	}
	gui.Run(cfg)
	return nil
}
