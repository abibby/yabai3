package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
	"github.com/abibby/yabai3/server"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	configPaths := []string{
		path.Join(cfg, "i3/config"),
		path.Join(home, ".config/i3/config"),
	}

	mainthread.Init(func() {
		var modeAST []*badparser.Mode
		var err error
		for _, path := range configPaths {
			modeAST, err = badparser.ParseFile(path)
			if !errors.Is(err, os.ErrNotExist) {
				break
			}
		}

		activeMode := "default"
		modes := map[string]*run.Mode{}

		changeMode := func(mode string) error {
			newMode, ok := modes[mode]
			if !ok {
				return fmt.Errorf("no mode %s", mode)
			}
			err := modes[activeMode].Unregister()
			if err != nil {
				return err
			}
			activeMode = mode
			log.Printf("Activate mode %s", mode)
			return newMode.Register()
		}

		go server.Serve(changeMode)

		done := make(chan struct{})
		for _, mode := range modeAST {
			m := run.NewMode()
			for _, b := range mode.BindSym {
				bind := b
				mods, key := keys(b.Keys)
				m.AddHotKey(mods, key, func(event hotkey.Event) {
					for _, c := range bind.Commands {
						err := run.Command(c, changeMode)
						if err != nil {
							log.Print(err)
						}
					}
				})
			}
			modes[mode.Name] = m

			if mode.Name == "default" {
				for _, w := range mode.Workspaces {
					err := run.LabelSpace(w.Displays, w.WorkspaceName)
					if err != nil {
						log.Print(err)
					}
				}
				err := run.SetGaps(mode.Borders.Inner, mode.Borders.Outer)
				if err != nil {
					log.Print(err)
				}
			}
		}

		modes[activeMode].Register()
		<-done
	})
}
