package run

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/abibby/yabai3/yabai"
)

var (
	ErrNoDisplay = fmt.Errorf("no display")
)

func cache[T any](timeout time.Duration, fetch func() (T, error)) func() (T, error) {
	var value T
	var hasValue = false
	var mtx = &sync.Mutex{}

	return func() (T, error) {
		mtx.Lock()
		defer mtx.Unlock()

		if !hasValue {
			v, err := fetch()
			if err != nil {
				var zero T
				return zero, err
			}
			value = v
			hasValue = true
		}
		return value, nil
	}
}

var getDisplays = cache(time.Second, yabai.QueryDisplays)

// var configuredSpaces = map[int]struct{}{}

func getDisplay(index int) (*yabai.Display, error) {
	displays, err := getDisplays()
	if err != nil {
		return nil, err
	}
	for _, d := range displays {
		if d.Index == index {
			return d, nil
		}
	}
	// spew.Dump(index, displays)
	return nil, ErrNoDisplay
}

func getDisplayFrom(displayNames []string) (*yabai.Display, error) {
	displays, err := getDisplays()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(displays, func(a, b *yabai.Display) int {
		return int(a.Frame.X) - int(b.Frame.X)
	})

	for _, name := range displayNames {
		switch name {
		case "left":
			return displays[0], nil
		case "center":
			return displays[len(displays)/2], nil
		case "right":
			return displays[len(displays)-1], nil
		}
	}

	return nil, ErrNoDisplay
}

func LabelSpace(displayNames []string, name string) error {
	d, err := getDisplayFrom(displayNames)
	if err != nil {
		return err
	}
	spaces, err := yabai.QuerySpaces()
	if err != nil {
		return err
	}

	for _, spaceIndex := range d.SpaceIndexes {
		for _, s := range spaces {
			if s.Label == name {
				return nil
			}
			if s.Label == "" {
				return yabai.Yabai("space", fmt.Sprint(spaceIndex), "--label", name)
			}
		}
	}
	return nil
}
