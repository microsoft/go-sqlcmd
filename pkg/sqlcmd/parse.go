// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strings"
	"unicode"
)

// grab grabs i from r, or returns 0 if i >= end.
func grab(r []rune, i, end int) rune {
	if i < end {
		return r[i]
	}
	return 0
}

// findNonSpace finds first non space rune in r, returning end if not found.
func findNonSpace(r []rune, i, end int) (int, bool) {
	for ; i < end; i++ {
		if !isSpaceOrControl(r[i]) {
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

// readCommand reads to the next control character to find
// a command in the string. Command regexes constrain matches
// to the beginning of the string, and all commands consume
// an entire line.
func readCommand(c Commands, r []rune, i, end int) (*Command, []string, int) {
	for ; i < end; i++ {
		next := grab(r, i, end)
		if next == 0 || unicode.IsControl(next) {
			break
		}
	}
	cmd, args := c.matchCommand(string(r[:i]))
	return cmd, args, i
}

// readVariableReference returns the length of the variable reference or false if it's not a valid identifier
func readVariableReference(r []rune, i int, end int) (int, bool) {
	for ; i < end; i++ {
		if r[i] == ')' {
			return i, true
		}
		if (r[i] >= 'a' && r[i] <= 'z') || (r[i] >= 'A' && r[i] <= 'Z') || (r[i] >= '0' && r[i] <= '9') || strings.ContainsRune(validVariableRunes, r[i]) {
			continue
		}
		break
	}
	return 0, false
}

func max64(a, b int64) int64 {
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

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// isSpaceOrControl is a special test for either a space or a control (ie, \b)
// characters.
func isSpaceOrControl(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}
