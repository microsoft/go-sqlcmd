// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tidwall/jsonc"
)

func TestPatchJSONCKey_PreservesComments(t *testing.T) {
	input := []byte(`{
  // Editor settings
  "editor.fontSize": 14,
  "editor.tabSize": 2,

  /* Database connections */
  "mssql.connections": [
    {
      "server": "old-server,1433",
      "profileName": "old-profile"
    }
  ],

  // Terminal settings
  "terminal.integrated.fontSize": 12
}`)

	newConns := []interface{}{
		map[string]interface{}{
			"server":      "new-server,1433",
			"profileName": "new-profile",
		},
	}

	result, err := patchJSONCKey(input, "mssql.connections", newConns)
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	s := string(result)

	if !strings.Contains(s, "// Editor settings") {
		t.Error("Line comment before key was destroyed")
	}
	if !strings.Contains(s, "/* Database connections */") {
		t.Error("Block comment before key was destroyed")
	}
	if !strings.Contains(s, "// Terminal settings") {
		t.Error("Line comment after key was destroyed")
	}
	if !strings.Contains(s, `"editor.fontSize": 14`) {
		t.Error("Other key was modified")
	}
	if !strings.Contains(s, `"terminal.integrated.fontSize": 12`) {
		t.Error("Other key was modified")
	}

	// Verify the patched file is valid JSONC (strip comments, then parse)
	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}

	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %v", m["mssql.connections"])
	}
	conn := conns[0].(map[string]interface{})
	if conn["server"] != "new-server,1433" {
		t.Errorf("Expected new-server, got %v", conn["server"])
	}
}

func TestPatchJSONCKey_InsertsNewKey(t *testing.T) {
	input := []byte(`{
  // Editor settings
  "editor.fontSize": 14
}`)

	newConns := []interface{}{
		map[string]interface{}{
			"server":      "localhost,1433",
			"profileName": "test",
		},
	}

	result, err := patchJSONCKey(input, "mssql.connections", newConns)
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, "// Editor settings") {
		t.Error("Comment was destroyed during insert")
	}
	if !strings.Contains(s, `"editor.fontSize": 14`) {
		t.Error("Existing key was modified during insert")
	}

	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}

	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %v", m["mssql.connections"])
	}
}

func TestPatchJSONCKey_EmptyFile(t *testing.T) {
	result, err := patchJSONCKey(nil, "mssql.connections", []interface{}{})
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Result is not valid JSON: %v\n%s", err, result)
	}
	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 0 {
		t.Errorf("Expected empty connections array, got %v", m["mssql.connections"])
	}
}

func TestPatchJSONCKey_TrailingComma(t *testing.T) {
	input := []byte(`{
  "editor.fontSize": 14,
  "mssql.connections": [],
}`)

	newConns := []interface{}{
		map[string]interface{}{"profileName": "test"},
	}

	result, err := patchJSONCKey(input, "mssql.connections", newConns)
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}
	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %v", m["mssql.connections"])
	}
}

func TestPatchJSONCKey_InsertIntoEmptyObject(t *testing.T) {
	input := []byte(`{}`)

	result, err := patchJSONCKey(input, "mssql.connections", []interface{}{})
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}
	if _, ok := m["mssql.connections"]; !ok {
		t.Error("Key was not inserted")
	}
}

func TestPatchJSONCKey_InlineCommentAfterValue(t *testing.T) {
	input := []byte(`{
  "mssql.connections": [] // old connections
}`)

	newConns := []interface{}{
		map[string]interface{}{"profileName": "new"},
	}

	result, err := patchJSONCKey(input, "mssql.connections", newConns)
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	if !strings.Contains(string(result), "// old connections") {
		t.Error("Inline comment after value was destroyed")
	}

	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}
}

func TestPatchJSONCKey_RealWorldVSCodeSettings(t *testing.T) {
	input := []byte(`{
  // General editor preferences
  "editor.fontSize": 14,
  "editor.wordWrap": "on",

  /*
   * Extensions
   */
  "extensions.autoUpdate": true,

  // SQL connections managed by sqlcmd
  "mssql.connections": [
    {
      "server": "localhost,1433",
      "profileName": "local-dev",
      "encrypt": "Optional",
      "trustServerCertificate": true,
    },
  ],

  // Python settings
  "python.linting.enabled": true,
  "python.formatting.provider": "black"
}`)

	newConns := []interface{}{
		map[string]interface{}{
			"server":                 "localhost,1433",
			"profileName":            "local-dev",
			"encrypt":                "Optional",
			"trustServerCertificate": true,
		},
		map[string]interface{}{
			"server":                 "prod-server,1433",
			"profileName":            "production",
			"encrypt":                "Mandatory",
			"trustServerCertificate": false,
			"authenticationType":     "SqlLogin",
			"user":                   "admin",
		},
	}

	result, err := patchJSONCKey(input, "mssql.connections", newConns)
	if err != nil {
		t.Fatalf("patchJSONCKey failed: %v", err)
	}

	s := string(result)

	// All comments preserved
	for _, comment := range []string{
		"// General editor preferences",
		"* Extensions",
		"// SQL connections managed by sqlcmd",
		"// Python settings",
	} {
		if !strings.Contains(s, comment) {
			t.Errorf("Comment %q was destroyed", comment)
		}
	}

	// All non-connection keys preserved verbatim
	for _, key := range []string{
		`"editor.fontSize": 14`,
		`"editor.wordWrap": "on"`,
		`"extensions.autoUpdate": true`,
		`"python.linting.enabled": true`,
		`"python.formatting.provider": "black"`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("Key %q was modified or lost", key)
		}
	}

	// Valid JSONC with correct data
	clean := jsonc.ToJSON(result)
	var m map[string]interface{}
	if err := json.Unmarshal(clean, &m); err != nil {
		t.Fatalf("Result is not valid JSONC: %v\n%s", err, result)
	}

	conns, ok := m["mssql.connections"].([]interface{})
	if !ok || len(conns) != 2 {
		t.Fatalf("Expected 2 connections, got %v", m["mssql.connections"])
	}
}
