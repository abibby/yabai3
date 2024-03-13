package run

import (
	"fmt"

	"golang.design/x/hotkey"
)

type Mode struct {
	hotkeys   []*hotkey.Hotkey
	callbacks map[*hotkey.Hotkey]func(hotkey.Event)
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

func (m *Mode) Register() error {
	for i, hk := range m.hotkeys {
		err := hk.Register()
		if err != nil {
			for _, hk2 := range m.hotkeys[:i] {
				hk2.Unregister()
			}
			return fmt.Errorf("%s: %w", hk, err)
		}
		go func(hk *hotkey.Hotkey) {
			for e := range hk.Keydown() {
				m.callbacks[hk](e)
			}
		}(hk)
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
