package tray

import (
	"context"
	"log"

	"github.com/abibby/salusa/di"
)

type VoidTray struct {
}

func (t *VoidTray) SetTitle(title string) {
}
func (t *VoidTray) SetTooltip(tooltip string) {
}
func (t *VoidTray) SetMenuItems(items []*MenuItem) {
}
func (t *VoidTray) Clicks() chan *MenuItem {
	return nil
}

func (t *VoidTray) Run(r Runnable) {
	defer func() {
		closeRunnable(r)
	}()
	err := r.Bootstrap()
	if err != nil {
		log.Fatal(err)
	}
	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func RegisterVoid(ctx context.Context) {
	di.RegisterSingleton(ctx, func() Tray {
		return &VoidTray{}
	})
}
