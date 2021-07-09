package sqlcmd

import "unicode"

// grab grabs i from r, or returns 0 if i >= end.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findSpace finds first space rune in r, returning end if not found.
func findSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if IsSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !IsSpaceOrControl(r[i]) {
			return i, true
		}
	}
	return i, false
}

// findRune finds the next rune c in r, returning end if not found.
func findRune(r []rune, i, end int, c rune) (int, bool) {
	for ; i < end; i++ {
		if r[i] == c {
			return i, true
		}
	}
	return i, false
}

// isEmptyLine returns true when r is empty or composed of only whitespace.
func isEmptyLine(r []rune, i, end int) bool {
	_, ok := findNonSpace(r, i, end)
	return !ok
}

// readString seeks to the end of a string returning the position and whether
// or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
func readString(r []rune, i, end int, quote rune) (int, bool) {
	var prev, c, next rune
	for ; i < end; i++ {
		c, next = r[i], grab(r, i+1, end)
		switch {
		case quote == '\'' && c == '\\':
			i++
			prev = 0
			continue
		case quote == '\'' && c == '\'' && next == '\'':
			i++
			continue
		case quote == '\'' && c == '\'' && prev != '\'',
			quote == '"' && c == '"':
			return i, true
		}
		prev = c
	}
	return end, false
}

// readMultilineComment finds the end of a multiline comment (ie, '*/').
func readMultilineComment(r []rune, i, end int) (int, bool) {
	i++
	for ; i < end; i++ {
		if r[i-1] == '*' && r[i] == '/' {
			return i, true
		}
	}
	return end, false
}

// Read to the next control character and try to find
// a command in the string. Command regexes constrain matches
// to the beginning of the string, and all commands consume
// an entire line.
func readCommand(r []rune, i, end int) (*Command, []string, int) {
	for ; i < end; i++ {
		next := grab(r, i, end)
		if next == 0 || unicode.IsControl(next) {
			break
		}
	}
	cmd, args := matchCommand(string(r[:i]))
	return cmd, args, i
}

// max returns the maximum of a, b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of a, b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsSpaceOrControl is a special test for either a space or a control (ie, \b)
// characters.
func IsSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

// RunesLastIndex returns the last index in r of needle, or -1 if not found.
func RunesLastIndex(r []rune, needle rune) int {
	i := len(r) - 1
	for ; i >= 0; i-- {
		if r[i] == needle {
			return i
		}
	}
	return i
}
