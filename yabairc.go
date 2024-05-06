package main

import (
	"log"

	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
	"github.com/abibby/yabai3/yabai"
)

func Yabairc() {
	modeAST, err := readConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var defaultMode *badparser.Mode
	for _, mode := range modeAST {
		if mode.Name == "default" {
			defaultMode = mode
		}
	}
	if defaultMode == nil {
		log.Fatal("no default mode")
	}
	// yabai -m signal --add event=dock_did_restart action="sudo yabai --load-sa"
	// sudo yabai --load-sa
	// yabai -m config layout bsp

	err = yabai.Yabai("config", "layout", "bsp")
	if err != nil {
		log.Print(err)
	}

	for _, w := range defaultMode.Workspaces {
		err := run.LabelSpace(w.DisplayIndexes, w.WorkspaceName)
		if err != nil {
			log.Print(err)
		}
	}
	err = run.SetGaps(defaultMode.Borders.Inner, defaultMode.Borders.Outer)
	if err != nil {
		log.Print(err)
	}
}
