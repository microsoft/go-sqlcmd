// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// patchJSONCKey replaces the value of a top-level key in a JSONC document,
// preserving all comments and formatting of other keys. If the key does not
// exist, it is appended before the closing brace. If data is nil or empty,
// a new JSON object is created.
func patchJSONCKey(data []byte, key string, value interface{}) ([]byte, error) {
	encoded, err := json.MarshalIndent(value, "  ", "  ")
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return []byte(fmt.Sprintf("{\n  %q: %s\n}\n", key, encoded)), nil
	}

	quotedKey := []byte(`"` + key + `"`)
	valStart, valEnd := findTopLevelJSONCValue(data, quotedKey)
	if valStart >= 0 {
		var buf bytes.Buffer
		buf.Write(data[:valStart])
		buf.Write(encoded)
		buf.Write(data[valEnd:])
		return buf.Bytes(), nil
	}

	return insertJSONCKey(data, quotedKey, encoded)
}

// findTopLevelJSONCValue locates a top-level key in JSONC data and returns
// the byte range [start, end) of its value. Returns (-1, -1) if not found.
func findTopLevelJSONCValue(data, quotedKey []byte) (int, int) {
	i := jsoncSkipWS(data, 0)
	if i >= len(data) || data[i] != '{' {
		return -1, -1
	}
	i++

	for {
		i = jsoncSkipWS(data, i)
		if i >= len(data) || data[i] == '}' {
			return -1, -1
		}
		if data[i] == ',' {
			i++
			continue
		}
		if data[i] != '"' {
			return -1, -1
		}

		keyStart := i
		i = jsoncSkipString(data, i)
		isTarget := bytes.Equal(data[keyStart:i], quotedKey)

		i = jsoncSkipWS(data, i)
		if i >= len(data) || data[i] != ':' {
			return -1, -1
		}
		i++

		i = jsoncSkipWS(data, i)
		valStart := i
		i = jsoncSkipValue(data, i)

		if isTarget {
			return valStart, i
		}
	}
}

// insertJSONCKey inserts a new key-value pair before the top-level closing brace.
func insertJSONCKey(data, quotedKey, encoded []byte) ([]byte, error) {
	var closingPos int
	i := jsoncSkipWS(data, 0)
	if i >= len(data) || data[i] != '{' {
		return nil, fmt.Errorf("no top-level object found")
	}
	depth := 0
	for i < len(data) {
		c := data[i]
		switch {
		case c == '"':
			i = jsoncSkipString(data, i)
			continue
		case i+1 < len(data) && c == '/' && data[i+1] == '/':
			i += 2
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		case i+1 < len(data) && c == '/' && data[i+1] == '*':
			i += 2
			for i+1 < len(data) {
				if data[i] == '*' && data[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		case c == '{' || c == '[':
			depth++
		case c == '}' || c == ']':
			depth--
			if depth == 0 && c == '}' {
				closingPos = i
				goto found
			}
		}
		i++
	}
	return nil, fmt.Errorf("no closing brace found")

found:
	needsComma := true
	for j := closingPos - 1; j >= 0; j-- {
		c := data[j]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if c == ',' || c == '{' {
			needsComma = false
		}
		break
	}

	var buf bytes.Buffer
	buf.Write(data[:closingPos])
	if needsComma {
		buf.WriteByte(',')
	}
	buf.WriteString("\n  ")
	buf.Write(quotedKey)
	buf.WriteString(": ")
	buf.Write(encoded)
	buf.WriteByte('\n')
	buf.Write(data[closingPos:])
	return buf.Bytes(), nil
}

// jsoncSkipWS advances past whitespace and JSONC comments.
func jsoncSkipWS(data []byte, i int) int {
	n := len(data)
	for i < n {
		c := data[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if i+1 < n && c == '/' {
			if data[i+1] == '/' {
				i += 2
				for i < n && data[i] != '\n' {
					i++
				}
				continue
			}
			if data[i+1] == '*' {
				i += 2
				for i+1 < n {
					if data[i] == '*' && data[i+1] == '/' {
						i += 2
						break
					}
					i++
				}
				continue
			}
		}
		break
	}
	return i
}

// jsoncSkipString advances past a quoted string including escape sequences.
func jsoncSkipString(data []byte, i int) int {
	n := len(data)
	if i >= n || data[i] != '"' {
		return i
	}
	i++
	for i < n {
		if data[i] == '\\' && i+1 < n {
			i += 2
			continue
		}
		if data[i] == '"' {
			return i + 1
		}
		i++
	}
	return i
}

// jsoncSkipValue advances past a JSONC value (string, number, object, array, bool, null).
func jsoncSkipValue(data []byte, i int) int {
	n := len(data)
	if i >= n {
		return i
	}
	switch data[i] {
	case '"':
		return jsoncSkipString(data, i)
	case '{', '[':
		return jsoncSkipBracket(data, i)
	default:
		for i < n {
			c := data[i]
			if c == ',' || c == '}' || c == ']' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				break
			}
			if c == '/' && i+1 < n && (data[i+1] == '/' || data[i+1] == '*') {
				break
			}
			i++
		}
		return i
	}
}

// jsoncSkipBracket advances past a bracket-delimited structure ({} or []),
// respecting nesting, strings, and comments.
func jsoncSkipBracket(data []byte, i int) int {
	n := len(data)
	if i >= n {
		return i
	}
	depth := 0
	for i < n {
		c := data[i]
		switch {
		case c == '"':
			i = jsoncSkipString(data, i)
			continue
		case i+1 < n && c == '/' && data[i+1] == '/':
			i += 2
			for i < n && data[i] != '\n' {
				i++
			}
			continue
		case i+1 < n && c == '/' && data[i+1] == '*':
			i += 2
			for i+1 < n {
				if data[i] == '*' && data[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		case c == '{' || c == '[':
			depth++
		case c == '}' || c == ']':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
		i++
	}
	return i
}
