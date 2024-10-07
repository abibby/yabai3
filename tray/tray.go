package tray

import (
	"context"
	"io"
	"log"

	"github.com/abibby/salusa/di"
)

type MenuItem struct {
	Title     string
	Tooltip   string
	Separator bool
	Clicks    chan *MenuItem
}

type Tray interface {
	SetTitle(title string)
	SetTooltip(tooltip string)
	SetMenuItems(items []*MenuItem)
	Run(r Runnable)
}

type Runnable interface {
	Bootstrap() error
	Run() error
}

func Run[T Runnable](ctx context.Context) {
	tray, err := di.Resolve[Tray](ctx)
	if err != nil {
		log.Fatal(err)
	}
	r, err := di.ResolveFill[T](ctx)
	if err != nil {
		log.Fatal(err)
	}
	tray.Run(r)
}

func closeRunnable(r Runnable) {
	if c, ok := r.(io.Closer); ok {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
