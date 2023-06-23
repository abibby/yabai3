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

func tokenizeLine(str string) []string {
	argv := newTokens()
	var quote rune
	escape := false
	for _, c := range str {
		if escape {
			switch c {
			case 'n':
				argv.Push('\n')
			case 't':
				argv.Push('\t')
			default:
				argv.Push(c)
			}
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
				argv.Next()
				continue
			}
			if c == '\\' {
				escape = true
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
