package run

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

type yabaiFrame struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"w"`
	Height float32 `json:"h"`
}

type yabaiWindow struct {
	ID                 int         `json:"id"`
	PID                int         `json:"pid"`
	App                string      `json:"app"`
	Title              string      `json:"title"`
	Frame              *yabaiFrame `json:"frame"`
	Role               string      `json:"role"`
	Subrole            string      `json:"subrole"`
	Display            int         `json:"display"`
	Space              int         `json:"space"`
	Level              int         `json:"level"`
	Opacity            float32     `json:"opacity"`
	SplitType          string      `json:"split-type"`
	SplitChild         string      `json:"split-child"`
	StackIndex         int         `json:"stack-index"`
	CanMove            bool        `json:"can-move"`
	CanResize          bool        `json:"can-resize"`
	HasFocus           bool        `json:"has-focus"`
	HasShadow          bool        `json:"has-shadow"`
	HasBorder          bool        `json:"has-border"`
	HasParentZoom      bool        `json:"has-parent-zoom"`
	HasFullscreenZoom  bool        `json:"has-fullscreen-zoom"`
	IsNativeFullscreen bool        `json:"is-native-fullscreen"`
	IsVisible          bool        `json:"is-visible"`
	IsMinimized        bool        `json:"is-minimized"`
	IsHidden           bool        `json:"is-hidden"`
	IsFloating         bool        `json:"is-floating"`
	IsSticky           bool        `json:"is-sticky"`
	IsTopmost          bool        `json:"is-topmost"`
	IsGrabbed          bool        `json:"is-grabbed"`
}

type yabaiSpace struct {
	ID                 int    `json:"id"`
	UUID               string `json:"uuid"`
	Index              int    `json:"index"`
	Label              string `json:"label"`
	Type               string `json:"type"`
	DisplayIndex       int    `json:"display"`
	WindowIDs          []int  `json:"windows"`
	FirstWindowID      int    `json:"first-window"`
	LastWindowID       int    `json:"last-window"`
	HasFocus           bool   `json:"has-focus"`
	IsVisible          bool   `json:"is-visible"`
	IsNativeFullscreen bool   `json:"is-native-fullscreen"`
}

type yabaiDisplay struct {
	ID           int         `json:"id"`
	UUID         string      `json:"uuid"`
	Index        int         `json:"index"`
	Frame        *yabaiFrame `json:"frame"`
	SpaceIndexes []int       `json:"spaces"`
}

func yabai(args ...string) error {
	return yabaiReturn(nil, args...)
}

func yabaiReturn(v any, args ...string) error {
	b, err := exec.Command("yabai", append([]string{"-m"}, args...)...).CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(b)))
	}
	if v == nil {
		return nil
	}

	return json.Unmarshal(b, v)
}

func yabaiQuerySpaces() ([]*yabaiSpace, error) {
	s := []*yabaiSpace{}
	err := yabaiReturn(&s, "query", "--spaces")
	if err != nil {
		return nil, err
	}
	return s, nil
}
func yabaiQueryActiveSpace() (*yabaiSpace, error) {
	s := &yabaiSpace{}
	err := yabaiReturn(s, "query", "--spaces", "--space")
	if err != nil {
		return nil, err
	}
	return s, nil
}

func yabaiQueryWindows() ([]*yabaiWindow, error) {
	w := []*yabaiWindow{}
	err := yabaiReturn(&w, "query", "--windows")
	if err != nil {
		return nil, err
	}
	return w, nil
}

func yabaiQueryActiveWindow() (*yabaiWindow, error) {
	w := &yabaiWindow{}
	err := yabaiReturn(&w, "query", "--windows", "--window")
	if err != nil {
		return nil, err
	}
	return w, nil
}

func yabaiQueryDisplays() ([]*yabaiDisplay, error) {
	d := []*yabaiDisplay{}
	err := yabaiReturn(&d, "query", "--displays")
	if err != nil {
		return nil, err
	}
	return d, nil
}
