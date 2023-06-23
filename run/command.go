package run

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var ErrUnknownCommand = errors.New("unknown command")

var directionMap = map[string]string{
	"up":    "north",
	"down":  "south",
	"left":  "west",
	"right": "east",
}

func Command(command []string) error {
	runners := map[string]func(c []string) error{
		"exec":      runExec,
		"focus":     runFocus,
		"move":      runMove,
		"resize":    runResize,
		"workspace": runWorkspace,
	}
	runner, ok := runners[command[0]]
	if !ok {
		return nil
	}
	err := runner(command)
	if err != nil {
		return fmt.Errorf("%s: %w", strings.Join(command, " "), err)
	}
	return nil
}

func runExec(c []string) error {
	argv := c[1:]

	b, err := exec.Command("open", append([]string{"-a", argv[0], "-n", "--args"}, argv[1:]...)...).CombinedOutput()
	if string(b) == fmt.Sprintf("Unable to find application named '%s'\n", argv[0]) {
		err = exec.Command(argv[0], argv[1:]...).Run()
	}
	if err != nil {
		return err
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

	err = yabai("window", "--resize", fmt.Sprintf("%s:%d:%d", direction, amount*horizontal, amount*vertial))
	if err == nil {
		return nil
	}

	amount = -amount
	if direction == "bottom" {
		direction = "top"
	} else if direction == "right" {
		direction = "left"
	}

	return yabai("window", "--resize", fmt.Sprintf("%s:%d:%d", direction, amount*horizontal, amount*vertial))

}
func runMove(c []string) error {
	direction, ok := directionMap[c[1]]
	if ok {
		err := yabai("window", "--swap", direction)
		if err == nil {
			return nil
		}
		nextSpace, err := getSpace(direction)
		if err != nil {
			return err
		}

		indexStr := fmt.Sprint(nextSpace.Index)

		return Command([]string{
			"move", "container", "to", "workspace", indexStr, ";",
			"workspace", indexStr,
		})
	}

	// workspacePrefix := "container to workspace"
	// if strings.HasPrefix(c.value, workspacePrefix) {
	// 	windowID := 0
	// 	if follow {
	// 		w, err := yabaiQueryActiveWindow()
	// 		if err != nil {
	// 			log.Print(err)
	// 		} else {
	// 			windowID = w.ID
	// 		}
	// 	}
	// 	workspace := unescapeString(c.value[len(workspacePrefix):])

	// 	err := yabai("window", "--space", workspace)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if windowID != 0 {
	// 		return yabai("window", "--focus", fmt.Sprint(windowID))
	// 	}
	// 	return nil
	// }

	return ErrUnknownCommand
}

func runFocus(c []string) error {
	direction, ok := directionMap[c[1]]
	if !ok {
		return ErrUnknownCommand
	}
	err := yabai("window", "--focus", direction)
	if err == nil {
		return nil
	}

	space, err := getSpace(direction)
	if err != nil {
		return err
	}

	return yabai("space", "--focus", fmt.Sprint(space.Index))
}

func runWorkspace(c []string) error {
	return yabai("space", "--focus", c[1])
}

func getSpace(direction string) (*yabaiSpace, error) {
	spaces, err := yabaiQuerySpaces()
	if err != nil {
		return nil, err
	}
	displays, err := yabaiQueryDisplays()
	if err != nil {
		return nil, err
	}
	var activeSpace *yabaiSpace
	var nextSpace *yabaiSpace
	var activeDisplay *yabaiDisplay
	var nextDisplay *yabaiDisplay

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
