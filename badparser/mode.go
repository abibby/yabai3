package badparser

import (
	"log"
	"strconv"
)

type Mode struct {
	Name       string
	BindSym    []*BindSym
	Workspaces []*Workspace
	Borders    *Borders
}

type BindSym struct {
	Keys     string
	Commands [][]string
}

type Workspace struct {
	WorkspaceName  string
	DisplayIndexes []string
}
type Borders struct {
	Inner int
	Outer int
}

func parseModes(modeLines map[string][]string) []*Mode {
	modes := []*Mode{}
	for name, lines := range modeLines {
		bindSym := []*BindSym{}
		windows := []*Workspace{}
		borders := &Borders{}
		for _, line := range lines {
			tokens := TokenizeLine(line)
			if len(tokens) == 0 {
				continue
			}
			switch tokens[0] {
			case "bindsym":
				bindSym = append(bindSym, &BindSym{
					Keys:     tokens[1],
					Commands: SplitCommands(tokens[2:]),
				})
			case "workspace":
				windows = append(windows, &Workspace{
					WorkspaceName:  tokens[1],
					DisplayIndexes: tokens[3:],
				})
			case "gaps":
				val, err := strconv.Atoi(tokens[2])
				if err != nil {
					log.Printf("parsing gaps: %v", err)
					continue
				}
				switch tokens[1] {
				case "inner":
					borders.Inner = val
				case "outer":
					borders.Outer = val
				}
			}
		}
		modes = append(modes, &Mode{
			Name:       name,
			BindSym:    bindSym,
			Workspaces: windows,
			Borders:    borders,
		})
	}
	return modes
}

func SplitCommands(tokens []string) [][]string {
	commands := [][]string{[]string{}}
	i := 0
	for _, t := range tokens {
		if t == ";" {
			i++
			commands = append(commands, []string{})
		} else {
			commands[i] = append(commands[i], t)
		}
	}
	return commands
}
