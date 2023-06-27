package badparser

import (
	"os"
	"strings"
)

func ParseFile(file string) ([]*Mode, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return Parse(string(b))
}

func Parse(src string) ([]*Mode, error) {
	lines := strings.Split(src, "\n")
	lines = trimWhitespace(lines)
	lines = replaceVariables(lines)
	modeLines := getModes(lines)
	modes := parseModes(modeLines)
	return modes, nil
}

func trimWhitespace(lines []string) []string {
	newLines := make([]string, len(lines))
	for i, line := range lines {
		newLines[i] = strings.TrimSpace(line)
	}
	return newLines
}

func replaceVariables(lines []string) []string {
	newLines := make([]string, len(lines))
	variables := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "set ") {
			parts := TokenizeLine(line)
			variables[parts[1]] = parts[2]
		}
	}

	for i, line := range lines {
		for name, value := range variables {
			line = strings.ReplaceAll(line, name, value)
		}
		newLines[i] = line
	}
	return newLines
}

func getModes(lines []string) map[string][]string {
	mode := "default"
	modes := map[string][]string{}
	addLine := func(line string) {
		m, ok := modes[mode]
		if !ok {
			m = []string{}
		}
		modes[mode] = append(m, line)
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "mode") {
			mode = TokenizeLine(line)[1]
		} else if line == "}" {
			mode = "default"
		}
		addLine(line)
	}
	return modes
}
