package run

import (
	"errors"
	"fmt"
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
		if err == nil || string(b) != fmt.Sprintf("Unable to find application named '%s'\n", args[0]) {
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
	if ok {
		err := yabai.Yabai("window", "--swap", direction)
		if err == nil {
			return nil
		}
		nextSpace, err := getSpaceInDirection(direction)
		if err != nil {
			return err
		}

		label := nextSpace.Label
		if label == "" {
			label = fmt.Sprint(nextSpace.Index)
		}

		return Command([]string{
			"move", "container", "to", "workspace", label, ";",
			"workspace", label,
		}, nil, nil)
	}

	if !slices.Equal([]string{"move", "container", "to", "workspace"}, c[:4]) {
		return ErrUnknownCommand
	}
	return yabai.Yabai("window", "--space", c[4])
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

	space, err := getSpaceInDirection(direction)
	if err != nil {
		return err
	}

	return yabai.Yabai("space", "--focus", fmt.Sprint(space.Index))
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

func getSpaceInDirection(direction string) (*yabai.Space, error) {
	spaces, err := yabai.QuerySpaces()
	if err != nil {
		return nil, err
	}
	displays, err := yabai.QueryDisplays()
	if err != nil {
		return nil, err
	}
	var activeSpace *yabai.Space
	var nextSpace *yabai.Space
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
		if d.Frame.Y == activeDisplay.Frame.Y {
			if direction == "east" && d.Frame.X == (activeDisplay.Frame.X+activeDisplay.Frame.Width) {
				nextDisplay = d
				break
			}
			if direction == "west" && d.Frame.X == (activeDisplay.Frame.X-activeDisplay.Frame.Width) {
				nextDisplay = d
				break
			}
		}
		if d.Frame.X == activeDisplay.Frame.X {
			if direction == "north" && d.Frame.Y == (activeDisplay.Frame.Y+activeDisplay.Frame.Height) {
				nextDisplay = d
				break
			}
			if direction == "south" && d.Frame.Y == (activeDisplay.Frame.Y-activeDisplay.Frame.Height) {
				nextDisplay = d
				break
			}
		}
	}

	if nextDisplay == nil {
		return nil, fmt.Errorf("no display %s of current display", direction)
	}

	for _, s := range spaces {
		if s.IsVisible && s.DisplayIndex == nextDisplay.Index {
			nextSpace = s
			break
		}
	}
	return nextSpace, nil
}
