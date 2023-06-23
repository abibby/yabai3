package badparser

type Mode struct {
	Name    string
	BindSym []*BindSym
}

func parseModes(modeLines map[string][]string) []*Mode {
	modes := []*Mode{}
	for name, lines := range modeLines {
		bindSym := []*BindSym{}
		for _, line := range lines {
			tokens := tokenizeLine(line)
			if len(tokens) == 0 {
				continue
			}
			if tokens[0] == "bindsym" {
				bindSym = append(bindSym, &BindSym{
					Keys:    tokens[1],
					Command: tokens[2:],
				})
			}
		}
		modes = append(modes, &Mode{
			Name:    name,
			BindSym: bindSym,
		})
	}
	return modes
}

type BindSym struct {
	Keys    string
	Command []string
}
