package badparser

import "strings"

type tokens struct {
	argv  []string
	index int
}

func newTokens() *tokens {
	return &tokens{
		argv: []string{""},
	}
}

func (t *tokens) Push(c rune) {
	t.argv[t.index] += string(c)
}
func (t *tokens) Next() {
	if t.argv[t.index] == "" {
		return
	}
	t.argv = append(t.argv, "")
	t.index++
}
func (t *tokens) Value() []string {
	if t.argv[t.index] == "" {
		return t.argv[0:t.index]
	}
	return t.argv
}

func TokenizeLine(str string) []string {
	argv := newTokens()
	var quote rune
	escape := 0
	for i, c := range str {
		if escape >= 2 {
			switch c {
			case 'n':
				argv.Push('\n')
			case 't':
				argv.Push('\t')
			default:
				argv.Push(c)
			}
			escape = 0
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
				argv.Next()
				continue
			}
			if c == '\\' {
				if escape == 1 {
					escape = 2
				} else if str[i+1] == '\\' {
					escape = 1
				}
				continue
			}
			argv.Push(c)
			continue
		}
		if c == '#' {
			return argv.Value()
		}
		if c == '"' || c == '\'' {
			quote = c
			argv.Next()
			continue
		}
		if strings.Contains("[]{}=;", string(c)) {
			argv.Next()
			argv.Push(c)
			argv.Next()
			continue
		}
		if strings.Contains(" \t", string(c)) {
			argv.Next()
			continue
		}
		argv.Push(c)
	}

	return argv.Value()
}
