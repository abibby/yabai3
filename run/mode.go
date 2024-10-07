package run

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.design/x/hotkey"
)

type Mode struct {
	hotkeys       []*hotkey.Hotkey
	callbacks     map[*hotkey.Hotkey]func(hotkey.Event)
	startup       []string
	startupAlways []string
}

func NewMode() *Mode {
	return &Mode{
		hotkeys:   []*hotkey.Hotkey{},
		callbacks: map[*hotkey.Hotkey]func(hotkey.Event){},
	}
}

func (m *Mode) AddHotKey(mods []hotkey.Modifier, key hotkey.Key, callback func(event hotkey.Event)) {
	hk := hotkey.New(mods, key)
	m.callbacks[hk] = callback
	m.hotkeys = append(m.hotkeys, hk)
}
func (m *Mode) SetStartup(s []string) {
	m.startup = s
}
func (m *Mode) SetStartupAlways(s []string) {
	m.startupAlways = s
}

func (m *Mode) Register() error {
	for i, hk := range m.hotkeys {
		err := hk.Register()
		if err != nil {
			registerErr := fmt.Errorf("key listener register %s: %w", hk, err)
			for _, hk2 := range m.hotkeys[:i] {
				err := hk2.Unregister()
				if err != nil {
					return errors.Join(err, registerErr)
				}
			}
			return registerErr
		}
		go func(hk *hotkey.Hotkey) {
			for e := range hk.Keydown() {
				m.callbacks[hk](e)
			}
		}(hk)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	for _, s := range m.startupAlways {
		err := exec.Command(shell, "-c", s).Run()
		if err != nil {
			log.Printf("exec_always failed: %s: %v", s, err)
		}
	}

	return nil
}

func (m *Mode) Unregister() error {
	for _, hk := range m.hotkeys {
		err := hk.Unregister()
		if err != nil {
			return err
		}
	}
	return nil
}
