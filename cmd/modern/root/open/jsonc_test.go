// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseJSONCSettings_Empty(t *testing.T) {
	m, err := parseJSONCSettings(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestParseJSONCSettings_StripsCommentsAndTrailingCommas(t *testing.T) {
	input := []byte(`{
  // line comment
  "key": "value", // inline comment
  /* block
     comment */
  "nums": [1, 2, 3,],
}`)
	m, err := parseJSONCSettings(input)
	if err != nil {
		t.Fatalf("parse failed: %v\ninput: %s", err, input)
	}
	if m["key"] != "value" {
		t.Errorf("expected key=value, got %v", m["key"])
	}
	if nums, ok := m["nums"].([]interface{}); !ok || len(nums) != 3 {
		t.Errorf("expected nums=[1,2,3], got %v", m["nums"])
	}
}

func TestParseJSONCSettings_StringsWithCommentLikeContent(t *testing.T) {
	input := []byte(`{
  "url": "http://example.com",
  "note": "has // slashes and /* stars */",
  "path": "C:\\Users\\test"
}`)
	m, err := parseJSONCSettings(input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if m["note"] != "has // slashes and /* stars */" {
		t.Errorf("string with comment-like content mangled: %v", m["note"])
	}
	if m["path"] != `C:\Users\test` {
		t.Errorf("escaped path mangled: %v", m["path"])
	}
}

func TestParseJSONCSettings_InvalidReturnsError(t *testing.T) {
	if _, err := parseJSONCSettings([]byte(`{"unterminated`)); err == nil {
		t.Error("expected error on invalid JSONC, got nil")
	}
}

func TestApplyJSONCSettingsUpdates_PreservesComments(t *testing.T) {
	original := []byte(`{
  // Editor settings
  "editor.fontSize": 14,
  "editor.tabSize": 2,

  /* mssql */
  "mssql.connections": [
    {"profileName": "old", "server": "old,1433"},
  ],

  // Terminal settings
  "terminal.integrated.fontSize": 12,
}`)
	updates := map[string]interface{}{
		"mssql.connections": []interface{}{
			map[string]interface{}{"profileName": "new", "server": "new,1433"},
		},
	}
	out, err := applyJSONCSettingsUpdates(original, updates)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	s := string(out)
	for _, want := range []string{
		"// Editor settings",
		"// Terminal settings",
		"/* mssql */",
		`"editor.fontSize": 14`,
		`"terminal.integrated.fontSize": 12`,
		`"profileName": "new"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, s)
		}
	}
	if strings.Contains(s, `"profileName": "old"`) {
		t.Errorf("old profile not replaced\nfull output:\n%s", s)
	}
}

func TestApplyJSONCSettingsUpdates_AddsMissingKeys(t *testing.T) {
	original := []byte(`{
  "editor.fontSize": 14,
  "editor.tabSize": 2
}`)
	updates := map[string]interface{}{
		"mssql.connections":      []interface{}{},
		"mssql.connectionGroups": []interface{}{},
	}
	out, err := applyJSONCSettingsUpdates(original, updates)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	m, err := parseJSONCSettings(out)
	if err != nil {
		t.Fatalf("re-parse failed: %v\noutput: %s", err, out)
	}
	for _, k := range []string{"editor.fontSize", "editor.tabSize", "mssql.connections", "mssql.connectionGroups"} {
		if _, ok := m[k]; !ok {
			t.Errorf("key %q missing after add\noutput: %s", k, out)
		}
	}
}

func TestApplyJSONCSettingsUpdates_EmptyOriginalReturnsFreshJSON(t *testing.T) {
	updates := map[string]interface{}{
		"mssql.connections": []interface{}{},
	}
	out, err := applyJSONCSettingsUpdates(nil, updates)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("fresh output is not valid JSON: %v\noutput: %s", err, out)
	}
	if _, ok := m["mssql.connections"]; !ok {
		t.Errorf("expected mssql.connections in fresh output, got %v", m)
	}
}

func TestJSONPointerEscape(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"mssql.connections", "mssql.connections"},
		{"a/b", "a~1b"},
		{"a~b", "a~0b"},
		{"a~/b", "a~0~1b"},
	}
	for _, c := range cases {
		if got := jsonPointerEscape(c.in); got != c.want {
			t.Errorf("escape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
