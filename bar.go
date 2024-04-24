package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/davecgh/go-spew/spew"
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

func runBar(command string) {

	cmd := exec.Command("sh", "-c", command)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()

	cmd.Stderr = NewPrefixWriter(os.Stderr, "bar | ")
	// cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start status command: %v", err)
	}
	defer func() {
		err := cmd.Wait()
		if err != nil {
			log.Fatalf("failed to stop bar: %v", err)
		}
	}()

	lastStatus := ""

	clicks := make(chan int)
	updates := make(chan Bar)
	go func() {
		s := bufio.NewScanner(stdout)
		s.Scan()
		s.Scan()
		for s.Scan() {
			bar := Bar{}
			err = json.Unmarshal(s.Bytes()[1:], &bar)
			if err != nil {
				log.Fatalf("error reading from bar: %v", err)
			}
			bar.Sort()
			updates <- bar
		}
	}()
	_, err = stdin.Write([]byte("[\n"))
	if err != nil {
		log.Printf("failed to write click header: %v", err)
		return
	}

	var bar Bar
	for {
		select {
		case clickIndex := <-clicks:
			section := bar[clickIndex]
			spew.Dump("click", section)
			b, err := json.Marshal(&Click{
				Name:      section.Name,
				Instance:  section.Instance,
				Button:    MouseLeft,
				Modifiers: []string{},
			})
			if err != nil {
				log.Printf("failed to encode click: %v", err)
				continue
			}
			err = writeAll(stdin, []byte(","), b, []byte("\n"))
			if err != nil {
				log.Printf("failed to send click: %v", err)
				continue
			}

		case bar = <-updates:
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

			newStatus := bar.String()
			if lastStatus != newStatus {
				systray.SetTitle(newStatus)
			}
			lastStatus = newStatus
		}
	}
}

type PrefixWriter struct {
	w         io.Writer
	prefix    string
	addPrefix bool
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{
		w:         w,
		prefix:    prefix,
		addPrefix: true,
	}
}

func (w *PrefixWriter) Write(b []byte) (n int, err error) {
	buff := make([]byte, 0, len(b)+len(w.prefix))
	for _, c := range b {
		if w.addPrefix {
			w.addPrefix = false
			buff = append(buff, []byte(w.prefix)...)
		}
		buff = append(buff, c)
		if c == '\n' {
			w.addPrefix = true
		}
	}
	return w.w.Write(buff)
}

func writeAll(w io.Writer, bytes ...[]byte) error {
	for _, b := range bytes {
		_, err := w.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}
