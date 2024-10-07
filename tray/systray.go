package tray

import (
	"context"
	"log"

	"github.com/abibby/salusa/di"
	"github.com/getlantern/systray"
)

type Systray struct {
	newItems     chan []*MenuItem
	clickIndexes chan int
}

func NewSystray() *Systray {
	s := &Systray{
		newItems:     make(chan []*MenuItem),
		clickIndexes: make(chan int, 4),
	}
	go s.handleChannels()
	return s
}

func (t *Systray) handleChannels() {
	var items []*MenuItem
	var systrayItems []*systray.MenuItem

	for {
		select {
		case items = <-t.newItems:
			systrayItems = t.setMenuItems(systrayItems, items)
		case i := <-t.clickIndexes:
			item := items[i]
			if item.Clicks != nil {
				item.Clicks <- items[i]
			}

		}
	}
}

func (t *Systray) SetTitle(title string) {
	systray.SetTitle(title)
}
func (t *Systray) SetTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}
func (t *Systray) SetMenuItems(items []*MenuItem) {
	t.newItems <- items
}
func (t *Systray) setMenuItems(systrayItems []*systray.MenuItem, items []*MenuItem) []*systray.MenuItem {
	for i, item := range items {
		if len(systrayItems) <= i {
			var m *systray.MenuItem
			if !item.Separator {
				m = systray.AddMenuItem(item.Title, item.Tooltip)
			} else {
				m = systray.AddMenuItem("---", "")
			}

			go func(menuIndex int) {
				for range m.ClickedCh {
					t.clickIndexes <- menuIndex
				}
			}(i)
			systrayItems = append(systrayItems, m)
		} else {
			systrayItems[i].Enable()
			systrayItems[i].SetTitle(item.Title)
			systrayItems[i].SetTooltip(item.Tooltip)
		}
	}
	for i := len(items); i < len(systrayItems); i++ {
		systrayItems[i].Disable()
	}
	return systrayItems
}
func (t *Systray) Close() error {
	close(t.clickIndexes)
	close(t.newItems)
	return nil
}
func (t *Systray) Run(r Runnable) {
	systray.Run(func() {
		err := r.Bootstrap()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			err := r.Run()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}, func() {
		closeRunnable(r)
		err := t.Close()
		if err != nil {
			log.Fatal(err)
		}
	})
}

func RegisterSystray(ctx context.Context) {
	di.RegisterSingleton(ctx, func() Tray {
		return NewSystray()
	})
}
