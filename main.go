package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func main() {
	mainthread.Init(func() {
		modes, err := badparser.ParseFile("./config")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		done := make(chan struct{})
		for _, mode := range modes {
			if mode.Name != "" {
				continue
			}
			for _, b := range mode.BindSym {
				_, err := registerHotKey(b)
				if err != nil {
					fmt.Fprintf(os.Stderr, "hotkey %s: %v\n", b.Keys, err)
				}
			}
		}
		<-done
	})
}

func registerHotKey(b *badparser.BindSym) (*hotkey.Hotkey, error) {
	hk := hotkey.New(keys(b.Keys))
	err := hk.Register()
	if err != nil {
		return nil, err
	}

	go func() {
		for range hk.Keydown() {
			err := run.Command(b.Command)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	return hk, nil
}
