package run

import (
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/abibby/yabai3/yabai"
	"github.com/mattn/go-shellwords"
	"golang.org/x/exp/slices"
)

var ErrUnknownCommand = errors.New("unknown command")

var directionMap = map[string]string{
	"up":    "north",
	"down":  "south",
	"left":  "west",
	"right": "east",
}

func Command(command []string, changeMode func(string) error, restart func() error) error {
	runners := map[string]func(c []string) error{
		"exec":       runExec,
		"focus":      runFocus,
		"move":       runMove,
		"resize":     runResize,
		"workspace":  runWorkspace,
		"mode":       runMode(changeMode),
		"fullscreen": runFullscreen,
		"restart":    runRestart(restart),
		"kill":       runKill,
	}
	runner, ok := runners[command[0]]
	if !ok {
		return fmt.Errorf("missing implementation for command %s", strings.Join(command, " "))
	}
	err := runner(command)
	if err != nil {
		return fmt.Errorf("%s: %w", strings.Join(command, " "), err)
	}
	return nil
}

func runExec(c []string) error {
	cmd := c[1]
	args, err := shellwords.Parse(cmd)
	if err == nil {
		b, err := exec.Command("open", append([]string{"-a", args[0], "-n", "--args"}, args[1:]...)...).CombinedOutput()
		if err == nil {
			return nil
		}
		if string(b) != fmt.Sprintf("Unable to find application named '%s'\n", args[0]) {
			return fmt.Errorf("exec failed with message: \"%s\": %w", b, err)
		}
	}
	b, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec failed with message: \"%s\": %w", b, err)
	}

	return nil
}

func runResize(c []string) error {
	parts := c[1:]
	horizontal := 0
	vertial := 0
	direction := "right"
	if parts[0] == "shrink" && parts[1] == "width" {
		direction = "right"
		horizontal = -1
	} else if parts[0] == "grow" && parts[1] == "width" {
		direction = "right"
		horizontal = 1
	} else if parts[0] == "shrink" && parts[1] == "height" {
		direction = "bottom"
		vertial = -1
	} else if parts[0] == "grow" && parts[1] == "height" {
		direction = "bottom"
		vertial = 1
	}

	amount, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}

	err = yabai.Yabai("window", "--resize", fmt.Sprintf("%s:%d:%d", direction, amount*horizontal, amount*vertial))
	if err == nil {
		return nil
	}

	amount = -amount
	if direction == "bottom" {
		direction = "top"
	} else if direction == "right" {
		direction = "left"
	}

	return yabai.Yabai("window", "--resize", fmt.Sprintf("%s:%d:%d", direction, amount*horizontal, amount*vertial))

}
func runMove(c []string) error {
	direction, ok := directionMap[c[1]]
	if !ok {
		if !slices.Equal([]string{"move", "container", "to", "workspace"}, c[:4]) {
			return ErrUnknownCommand
		}
		return yabai.Yabai("window", "--space", c[4])
	}

	err := yabai.Yabai("window", "--swap", direction)
	if err == nil {
		return nil
	}

	window, err := yabai.QueryActiveWindow()
	if err != nil {
		return err
	}
	nextSpace, err := getSpaceInDirection(direction)
	if err != nil {
		return err
	}

	label := nextSpace.Label
	if label == "" {
		label = fmt.Sprint(nextSpace.Index)
	}

	err = yabai.Yabai("window", "--space", label)
	if err != nil {
		return err
	}

	return yabai.Yabai("window", "--focus", fmt.Sprint(window.ID))
}

func runFocus(c []string) error {
	direction, ok := directionMap[c[1]]
	if !ok {
		return ErrUnknownCommand
	}
	err := yabai.Yabai("window", "--focus", direction)
	if err == nil {
		return nil
	}

	display, err := getDisplayInDirection(direction)
	if err != nil {
		return err
	}

	return yabai.Yabai("display", "--focus", fmt.Sprint(display.Index))
}

func runWorkspace(c []string) error {
	return yabai.Yabai("space", "--focus", c[1])
}

func runMode(changeMode func(string) error) func(c []string) error {
	return func(c []string) error {
		return changeMode(c[1])
	}
}

func runFullscreen(c []string) error {
	if c[1] == "toggle" {
		return yabai.Yabai("window", "--toggle", "zoom-fullscreen")
	}
	return ErrUnknownCommand
}

func runRestart(restart func() error) func(c []string) error {
	return func(c []string) error {
		return restart()
	}
}
func runKill(c []string) error {
	w, err := yabai.QueryActiveWindow()
	if err != nil {
		return err
	}
	return syscall.Kill(w.PID, syscall.SIGTERM)
}
func findAngleBetween(d1, d2 *yabai.Display) float64 {
	x1 := d1.Frame.X + d1.Frame.Width/2
	y1 := d1.Frame.Y + d1.Frame.Height/2
	x2 := d2.Frame.X + d2.Frame.Width/2
	y2 := d2.Frame.Y + d2.Frame.Height/2
	calcAngle := math.Atan2(float64(y2-y1), float64(x2-x1))
	if calcAngle < 0 {
		calcAngle += math.Pi * 2
	}
	return calcAngle * (180 / math.Pi)
}
func getDisplayInDirection(direction string) (*yabai.Display, error) {
	spaces, err := yabai.QuerySpaces()
	if err != nil {
		return nil, err
	}
	displays, err := yabai.QueryDisplays()
	if err != nil {
		return nil, err
	}
	var activeSpace *yabai.Space
	var activeDisplay *yabai.Display
	var nextDisplay *yabai.Display

	for _, s := range spaces {
		if s.HasFocus {
			activeSpace = s
			break
		}
	}

	for _, d := range displays {
		if d.Index == activeSpace.DisplayIndex {
			activeDisplay = d
			break
		}
	}
	for _, d := range displays {
		if d.ID == activeDisplay.ID {
			continue
		}
		// TODO: look into using the closet display within an angle range
		// angle := findAngleBetween(activeDisplay, d)
		// if d.Frame.Y == activeDisplay.Frame.Y {
		if direction == "east" && d.Frame.X == (activeDisplay.Frame.X+activeDisplay.Frame.Width) {
			nextDisplay = d
			break
		}

		if direction == "west" && activeDisplay.Frame.X == (d.Frame.X+d.Frame.Width) {
			nextDisplay = d
			break
		}
		// }
		// if d.Frame.X == activeDisplay.Frame.X {
		if direction == "north" && d.Frame.Y == (activeDisplay.Frame.Y+activeDisplay.Frame.Height) {
			nextDisplay = d
			break
		}
		if direction == "south" && activeDisplay.Frame.Y == (d.Frame.Y+d.Frame.Height) {
			nextDisplay = d
			break
		}
		// }
	}

	if nextDisplay == nil {
		return nil, fmt.Errorf("no display %s of current display", direction)
	}
	return nextDisplay, nil
}
func getSpaceInDirection(direction string) (*yabai.Space, error) {
	var nextSpace *yabai.Space
	spaces, err := yabai.QuerySpaces()
	if err != nil {
		return nil, err
	}
	nextDisplay, err := getDisplayInDirection(direction)
	if err != nil {
		return nil, err
	}
	for _, s := range spaces {
		if s.IsVisible && s.DisplayIndex == nextDisplay.Index {
			nextSpace = s
			break
		}
	}
	return nextSpace, nil
}
