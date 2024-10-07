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

	"github.com/abibby/salusa/di"
	"github.com/abibby/yabai3/tray"
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

func Run(ctx context.Context, command string, globalMenuItems []*tray.MenuItem) {
	t, err := di.Resolve[tray.Tray](ctx)
	if err != nil {
		log.Fatal(err)
	}
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

	clicks := make(chan *tray.MenuItem)
	updates := Process(stdout)

	_, err = stdin.Write([]byte("[\n"))
	if err != nil {
		log.Printf("failed to write click header: %v", err)
		return
	}

	var bar Bar
	for {
		select {
		case <-ctx.Done():
			return
		case clicked := <-clicks:
			i := slices.Index(globalMenuItems, clicked)
			if i == -1 {
				continue
			}
			err = doClick(bar[i], stdin)
			if err != nil {
				log.Print(err)
			}
		case bar = <-updates:
			status := barUpdate(t, bar, globalMenuItems, clicks)
			if lastStatus != status {
				log.Printf("bar update %s\n", status)
				t.SetTitle(status)
				lastStatus = status
			}
		}
	}
}

func barUpdate(t tray.Tray, bar Bar, globalMenuItems []*tray.MenuItem, clicks chan *tray.MenuItem) string {
	activeSections := bar.ActiveSections()
	offset := len(globalMenuItems)
	items := make([]*tray.MenuItem, offset, offset+len(activeSections))
	copy(items, globalMenuItems)
	for _, section := range activeSections {
		items = append(items, &tray.MenuItem{
			Title:   section.String(),
			Tooltip: section.Name,
			Clicks:  clicks,
		})
	}

	t.SetMenuItems(items)

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
