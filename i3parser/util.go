package i3parser

func isWhitespace(b byte) bool {
	return isInlineWhitespace(b) || b == '\n'
}

func isInlineWhitespace(b byte) bool {
	return b == ' ' || b == '\t'
}

// func ParseVariableName(block *parser.Reader) (string, error) {
// 	block.SkipWhitespace()
// 	word := block.PeakWord()
// 	if len(word) == 0 {
// 		return "", fmt.Errorf("cound not find identifier")
// 	}
// 	if word[0] != '$' {
// 		return "", fmt.Errorf("invalid variable name %s: must start with $", word)
// 	}
// 	block.Advance(len(word))
// 	return string(word), nil
// }
