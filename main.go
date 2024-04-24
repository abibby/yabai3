package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
	"github.com/abibby/yabai3/server"
	"github.com/getlantern/systray"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

type State uint8

const (
	Running = State(iota)
	Restart
	Stopped
)

func main() {
	// go run()

	systray.Run(onReady, onExit)
}

func onExit() {
	// clean up here
}
func onReady() {
	systray.SetTitle("yabai3")
	systray.SetTooltip("yabai3")
	mQuit := systray.AddMenuItem("Quit", "Quit yabai3")
	mRestart := systray.AddMenuItem("Restart", "Restart yabai and yabai3")
	systray.AddSeparator()

	done := make(chan State)

	go func() {
		select {
		case <-mQuit.ClickedCh:
			done <- Stopped
		case <-mRestart.ClickedCh:
			done <- Restart
		}

	}()

	mainthread.Init(func() {
		state := Running
		for state != Stopped {
			log.Print("starting yabai3")
			var modeAST []*badparser.Mode
			var err error
			modeAST, err = readConfig()
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
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

			restart := func() error {
				err := exec.Command("yabai", "--restart-service").Run()
				if err != nil {
					return fmt.Errorf("faild to restart yabai: %w", err)
				}
				log.Print("restarted yabai service")
				time.Sleep(time.Second * 5)
				done <- Restart
				return nil
			}

			stopServer := server.Serve(changeMode, restart)

			for _, mode := range modeAST {
				m := run.NewMode()
				for _, b := range mode.BindSym {
					bind := b
					mods, key := keys(b.Keys)
					m.AddHotKey(mods, key, func(event hotkey.Event) {
						for _, c := range bind.Commands {
							err := run.Command(c, changeMode, restart)
							if err != nil {
								log.Print(err)
							}
						}
					})
				}
				modes[mode.Name] = m

				if mode.Name == "default" {
					for _, w := range mode.Workspaces {
						err := run.LabelSpace(w.DisplayIndexes, w.WorkspaceName)
						if err != nil {
							log.Print(err)
						}
					}
					err := run.SetGaps(mode.Borders.Inner, mode.Borders.Outer)
					if err != nil {
						log.Print(err)
					}
					go runBar(mode.Bar.StatusCommand)
				}
			}

			err = modes[activeMode].Register()
			if err != nil {
				log.Printf("failed to register bindings: %v", err)
				return
			}
			log.Print("listening for key bindings")
			state = <-done
			err = modes[activeMode].Unregister()
			if err != nil {
				log.Printf("failed to unregister bindings: %v", err)
				return
			}
			stopServer()
		}
	})
}

func readConfig() ([]*badparser.Mode, error) {
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

	for _, path := range configPaths {
		modeAST, err := badparser.ParseFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		return modeAST, nil
	}
	return nil, fmt.Errorf("no config file found")
}
