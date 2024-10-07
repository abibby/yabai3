package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	// _ "net/http/pprof"

	"github.com/abibby/salusa/di"
	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/bar"
	"github.com/abibby/yabai3/run"
	"github.com/abibby/yabai3/server"
	"github.com/abibby/yabai3/tray"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

var (
	ErrStop    = errors.New("stop")
	ErrRestart = errors.New("restart")
)

func main() {
	var command string
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	switch command {
	case "yabairc":
		Yabairc()
	default:
		ctx := di.ContextWithDependencyProvider(
			context.Background(),
			di.NewDependencyProvider(),
		)
		tray.RegisterVoid(ctx)
		// tray.RegisterSystray(ctx)
		tray.Run[*Service](ctx)
	}
}

type Service struct {
	Tray tray.Tray       `inject:""`
	Ctx  context.Context `inject:""`

	menuItems []*tray.MenuItem
	clicks    chan *tray.MenuItem
}

func (s *Service) Bootstrap() error {
	log.Print("Starting yabai3 0.1.0")
	s.Tray.SetTitle("yabai3")
	s.Tray.SetTooltip("yabai3")

	s.clicks = make(chan *tray.MenuItem)
	s.menuItems = []*tray.MenuItem{
		{Title: "Quit", Tooltip: "Quit yabai3", Clicks: s.clicks},
		{Title: "Restart", Tooltip: "Restart yabai and yabai3", Clicks: s.clicks},
		{Separator: true},
	}
	return nil
}
func (s *Service) Run() error {
	mainthread.Init(func() {
		cause := ErrRestart
		for cause == ErrRestart {
			ctx, cancel := context.WithCancelCause(s.Ctx)
			go func() {
				select {
				case <-ctx.Done():
					return
				case m := <-s.clicks:
					if m.Title == "Quit" {
						cancel(ErrStop)
					} else if m.Title == "Restart" {
						cancel(ErrRestart)
					}
				}

			}()
			err := s.do(ctx, cancel)
			if err != nil {
				panic(err)
			}
			cause = context.Cause(ctx)
		}
		if cause != ErrStop {
			panic(cause)
		}
	})
	return nil
}

func (s *Service) do(ctx context.Context, cancel context.CancelCauseFunc) error {
	log.Print("starting yabai3")
	var modeAST []*badparser.Mode
	var err error
	modeAST, err = readConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	activeMode := "default"
	modes := map[string]*run.Mode{}

	i3MsgServer := server.New()

	changeMode := func(mode string) error {
		i3MsgServer.ModeChanged(mode)
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
		cancel(ErrRestart)
		return nil
	}

	err = i3MsgServer.Start(ctx, changeMode, restart)
	if err != nil {
		return err
	}
	defer func() {
		err := i3MsgServer.Close()
		if err != nil {
			log.Printf("failed to unregister bindings: %v", err)
		}
	}()

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

		if mode.Name == "default" {
			go bar.Run(ctx, mode.Bar.StatusCommand, s.menuItems)

			m.SetStartup(mode.Exec)
			m.SetStartupAlways(mode.ExecAlways)
		}
		modes[mode.Name] = m
	}

	err = modes[activeMode].Register()
	if err != nil {
		return fmt.Errorf("failed to register bindings: %w", err)
	}
	log.Print("listening for key bindings")

	<-ctx.Done()

	err = modes[activeMode].Unregister()
	if err != nil {
		return fmt.Errorf("failed to unregister bindings: %w", err)
	}

	return nil
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
