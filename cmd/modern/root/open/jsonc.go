// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/tailscale/hujson"
)

// parseJSONCSettings parses a JSONC document into a generic map.
// Empty input returns an empty map.
func parseJSONCSettings(data []byte) (map[string]interface{}, error) {
	settings := make(map[string]interface{})
	if len(bytes.TrimSpace(data)) == 0 {
		return settings, nil
	}
	v, err := hujson.Parse(data)
	if err != nil {
		return nil, err
	}
	v.Standardize()
	if err := json.Unmarshal(v.Pack(), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// applyJSONCSettingsUpdates sets the given top-level keys in original via an
// RFC 6902 patch on the hujson AST, leaving comments, trailing commas, and
// unrelated keys intact. Empty original yields a fresh JSON document.
func applyJSONCSettingsUpdates(original []byte, updates map[string]interface{}) ([]byte, error) {
	if len(bytes.TrimSpace(original)) == 0 {
		out, err := json.MarshalIndent(updates, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(out, '\n'), nil
	}

	v, err := hujson.Parse(original)
	if err != nil {
		return nil, err
	}

	std := v.Clone()
	std.Standardize()
	var existing map[string]json.RawMessage
	if err := json.Unmarshal(std.Pack(), &existing); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	type patchOp struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
	ops := make([]patchOp, 0, len(keys))
	for _, k := range keys {
		raw, err := json.Marshal(updates[k])
		if err != nil {
			return nil, err
		}
		op := "add"
		if _, ok := existing[k]; ok {
			op = "replace"
		}
		ops = append(ops, patchOp{Op: op, Path: "/" + jsonPointerEscape(k), Value: raw})
	}

	patch, err := json.Marshal(ops)
	if err != nil {
		return nil, err
	}
	if err := v.Patch(patch); err != nil {
		return nil, err
	}
	// Format re-indents inserted members; it preserves comments.
	v.Format()
	return v.Pack(), nil
}

// jsonPointerEscape escapes a JSON Pointer reference token per RFC 6901.
func jsonPointerEscape(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}
