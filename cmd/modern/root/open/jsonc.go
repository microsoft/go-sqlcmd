// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

// stripJSONC removes comments (// and /* */) and trailing commas from JSONC
// data, producing valid JSON. String literals are preserved as-is.
func stripJSONC(data []byte) []byte {
	var result []byte
	i := 0
	n := len(data)

	for i < n {
		// String literal: copy verbatim, respecting escape sequences
		if data[i] == '"' {
			result = append(result, data[i])
			i++
			for i < n && data[i] != '"' {
				if data[i] == '\\' && i+1 < n {
					result = append(result, data[i], data[i+1])
					i += 2
					continue
				}
				result = append(result, data[i])
				i++
			}
			if i < n {
				result = append(result, data[i]) // closing "
				i++
			}
			continue
		}

		// Line comment: skip to end of line
		if i+1 < n && data[i] == '/' && data[i+1] == '/' {
			i += 2
			for i < n && data[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment: skip to closing */
		if i+1 < n && data[i] == '/' && data[i+1] == '*' {
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

		result = append(result, data[i])
		i++
	}

	// Second pass: remove trailing commas before ] or }
	cleaned := make([]byte, 0, len(result))
	for i := 0; i < len(result); i++ {
		if result[i] == ',' {
			j := i + 1
			for j < len(result) && (result[j] == ' ' || result[j] == '\t' || result[j] == '\n' || result[j] == '\r') {
				j++
			}
			if j < len(result) && (result[j] == ']' || result[j] == '}') {
				continue // skip trailing comma
			}
		}
		cleaned = append(cleaned, result[i])
	}

	return cleaned
}
