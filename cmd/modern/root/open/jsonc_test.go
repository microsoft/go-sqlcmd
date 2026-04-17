// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"testing"
)

func TestStripJSONC_LineComments(t *testing.T) {
	input := []byte(`{
  // This is a comment
  "key": "value" // inline comment
}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse stripped JSONC: %v\nResult: %s", err, result)
	}
	if m["key"] != "value" {
		t.Errorf("Expected 'value', got %v", m["key"])
	}
}

func TestStripJSONC_BlockComments(t *testing.T) {
	input := []byte(`{
  /* block comment */
  "key": "value",
  /*
   * multi-line
   * block comment
   */
  "other": 42
}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse stripped JSONC: %v\nResult: %s", err, result)
	}
	if m["key"] != "value" {
		t.Errorf("Expected 'value', got %v", m["key"])
	}
	if m["other"] != float64(42) {
		t.Errorf("Expected 42, got %v", m["other"])
	}
}

func TestStripJSONC_TrailingCommas(t *testing.T) {
	input := []byte(`{
  "a": 1,
  "b": [1, 2, 3,],
  "c": {"x": 1,},
}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse stripped JSONC: %v\nResult: %s", err, result)
	}
	if m["a"] != float64(1) {
		t.Errorf("Expected 1, got %v", m["a"])
	}
}

func TestStripJSONC_CommentsInStringsPreserved(t *testing.T) {
	input := []byte(`{
  "url": "http://example.com",
  "note": "has // slashes and /* stars */",
  "path": "C:\\Users\\test"
}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse stripped JSONC: %v\nResult: %s", err, result)
	}
	if m["url"] != "http://example.com" {
		t.Errorf("URL mangled: %v", m["url"])
	}
	if m["note"] != "has // slashes and /* stars */" {
		t.Errorf("String with comment-like content mangled: %v", m["note"])
	}
	if m["path"] != `C:\Users\test` {
		t.Errorf("Escaped path mangled: %v", m["path"])
	}
}

func TestStripJSONC_RealWorldVSCodeSettings(t *testing.T) {
	// Realistic VS Code settings.json with JSONC features
	input := []byte(`{
  // Editor settings
  "editor.fontSize": 14,
  "editor.tabSize": 2,

  /* Database connections */
  "mssql.connections": [
    {
      "server": "localhost,1433",
      "profileName": "my-db",
      "encrypt": "Optional",
      "trustServerCertificate": true,
    },
  ],

  // Terminal settings
  "terminal.integrated.fontSize": 12,
}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse real-world JSONC: %v\nResult: %s", err, result)
	}
	if m["editor.fontSize"] != float64(14) {
		t.Errorf("Expected fontSize 14, got %v", m["editor.fontSize"])
	}
	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %v", m["mssql.connections"])
	}
}

func TestStripJSONC_EmptyInput(t *testing.T) {
	result := stripJSONC([]byte{})
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %s", result)
	}
}

func TestStripJSONC_PureJSON(t *testing.T) {
	// No comments, no trailing commas - should pass through cleanly
	input := []byte(`{"key": "value", "num": 42}`)
	result := stripJSONC(input)
	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to parse pure JSON: %v", err)
	}
	if m["key"] != "value" || m["num"] != float64(42) {
		t.Errorf("Values changed: %v", m)
	}
}
