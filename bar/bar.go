package bar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"slices"
	"strings"

	"github.com/getlantern/systray"
	"github.com/microcosm-cc/bluemonday"
)

type BarSection struct {
	id       int
	Name     string `json:"Name"`
	Instance string `json:"Instance"`
	Text     string `json:"full_text"`
	Markup   string `json:"markup,omitempty"`
}

func (b *BarSection) String() string {
	return bluemonday.StrictPolicy().Sanitize(b.Text)
}

type Bar []*BarSection

func (b *Bar) Sort() {
	slices.SortFunc(*b, func(a, b *BarSection) int {
		return a.id - b.id
	})
}

func (b *Bar) String() string {
	parts := []string{}
	for _, sec := range *b {
		if sec.Text != "" {
			parts = append(parts, sec.String())
		}
	}
	return strings.Join(parts, " | ")
}

func (b *Bar) ActiveSections() []*BarSection {
	sections := []*BarSection{}
	for _, sec := range *b {
		if sec.Text != "" {
			sections = append(sections, sec)
		}
	}
	return sections
}

type MouseButton uint8

const (
	MouseLeft       = 1
	MouseMiddle     = 2
	MouseRight      = 3
	MouseScrollUp   = 4
	MouseScrollDown = 5
	MouseBack       = 8
	MouseForward    = 9
)

type Click struct {
	Name      string      `json:"name"`
	Instance  string      `json:"instance"`
	Button    MouseButton `json:"button"`
	Modifiers []string    `json:"modifiers"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	RelativeX int         `json:"relative_x"`
	RelativeY int         `json:"relative_y"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
}

var menuItems = []*systray.MenuItem{}

func Run(ctx context.Context, command string) {
	stdin, stdout, cmd := StartCommand(ctx, command)
	defer stdin.Close()
	defer stdout.Close()

	defer func() {
		err := cmd.Wait()
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			log.Printf("bar exited with code: %d", exitErr.ExitCode())
			return
		} else if err != nil {
			log.Fatalf("failed to stop bar: %v", err)
		}
	}()

	lastStatus := ""

	clicks := make(chan int)
	updates := Process(stdout)

	_, err := stdin.Write([]byte("[\n"))
	if err != nil {
		log.Printf("failed to write click header: %v", err)
		return
	}

	var bar Bar
	for {
		select {
		case <-ctx.Done():
			return
		case clickIndex := <-clicks:
			err = doClick(bar[clickIndex], stdin)
			if err != nil {
				log.Print(err)
			}
		case bar = <-updates:
			status := barUpdate(bar, clicks)
			if lastStatus != status {
				log.Printf("bar update %s\n", status)
				systray.SetTitle(status)
				lastStatus = status
			}
		}
	}
}

func barUpdate(bar Bar, clicks chan int) string {
	activeSections := bar.ActiveSections()
	for i, section := range activeSections {
		title := section.String()
		tooltip := section.Name
		if len(menuItems) <= i {
			m := systray.AddMenuItem(title, tooltip)
			go func(menuIndex int) {
				for {
					<-m.ClickedCh
					clicks <- menuIndex
				}
			}(i)
			menuItems = append(menuItems, m)
		} else {
			menuItems[i].Enable()
			menuItems[i].SetTitle(title)
			menuItems[i].SetTooltip(tooltip)
		}
	}
	for i := len(activeSections); i < len(menuItems); i++ {
		menuItems[i].Disable()
	}

	return bar.String()
}

func doClick(section *BarSection, w io.Writer) error {
	b, err := json.Marshal(&Click{
		Name:      section.Name,
		Instance:  section.Instance,
		Button:    MouseLeft,
		Modifiers: []string{},
	})
	if err != nil {
		return fmt.Errorf("failed to encode click: %v", err)
	}
	err = writeAll(w, []byte(","), b, []byte("\n"))
	if err != nil {
		return fmt.Errorf("failed to send click: %v", err)
	}
	return nil
}
