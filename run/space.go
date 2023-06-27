package run

import (
	"fmt"
	"strconv"
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
var getSpaces = cache(time.Second, yabai.QuerySpaces)

var configuredSpaces = map[int]struct{}{}

func getDisplay(id int) (*yabai.Display, error) {
	displays, err := getDisplays()
	if err != nil {
		return nil, err
	}
	for _, d := range displays {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, ErrNoDisplay
}
func getDisplayFrom(displayIDs []string) (*yabai.Display, error) {
	for _, strID := range displayIDs {
		id, err := strconv.Atoi(strID)
		if err != nil {
			continue
		}
		d, err := getDisplay(id)
		if err == ErrNoDisplay {
			continue
		}
		return d, err
	}
	return nil, ErrNoDisplay
}
func getSpace(id int) (*yabai.Space, error) {
	spaces, err := getSpaces()
	if err != nil {
		return nil, err
	}
	for _, s := range spaces {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, ErrNoDisplay
}

func LabelSpace(displayIDs []string, name string) error {
	d, err := getDisplayFrom(displayIDs)
	if err != nil {
		return err
	}
	for _, spaceIndex := range d.SpaceIndexes {
		_, ok := configuredSpaces[spaceIndex]
		if ok {
			continue
		}
		configuredSpaces[spaceIndex] = struct{}{}
		return yabai.Yabai("space", fmt.Sprint(spaceIndex), "--label", name)
	}
	return nil
}
