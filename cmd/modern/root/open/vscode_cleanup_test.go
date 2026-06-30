// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveContextFromVSCodeSettings(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	testSettingsPathOverride = path
	t.Cleanup(func() { testSettingsPathOverride = "" })

	initial := `{
    // user comment that must survive
    "editor.fontSize": 14,
    "mssql.connections": [
        { "profileName": "keep-me",   "server": "kept,1433" },
        { "profileName": "doomed-ctx", "server": "localhost,1433", "password": "shh" }
    ]
}
`
	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}

	cleaned := RemoveContextFromVSCodeSettings("doomed-ctx")
	if len(cleaned) != 1 || cleaned[0] != path {
		t.Fatalf("expected cleanup to report %s, got %v", path, cleaned)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "user comment that must survive") {
		t.Errorf("comment was stripped: %s", data)
	}

	settings, err := parseJSONCSettings(data)
	if err != nil {
		t.Fatalf("settings no longer parse: %v\n%s", err, data)
	}

	conns, _ := settings["mssql.connections"].([]interface{})
	if len(conns) != 1 {
		t.Fatalf("expected exactly 1 remaining connection, got %d: %s", len(conns), mustJSON(conns))
	}
	name, _ := conns[0].(map[string]interface{})["profileName"].(string)
	if name != "keep-me" {
		t.Errorf("wrong connection survived: %s", name)
	}

	// Second call is a no-op once the entry is gone.
	if cleaned := RemoveContextFromVSCodeSettings("doomed-ctx"); len(cleaned) != 0 {
		t.Errorf("expected no cleanup on second call, got %v", cleaned)
	}
}

func TestRemoveContextFromVSCodeSettings_MissingFile(t *testing.T) {
	testSettingsPathOverride = filepath.Join(t.TempDir(), "does-not-exist.json")
	t.Cleanup(func() { testSettingsPathOverride = "" })

	if cleaned := RemoveContextFromVSCodeSettings("anything"); len(cleaned) != 0 {
		t.Errorf("expected no cleanup when file is missing, got %v", cleaned)
	}
}

func TestRemoveContextFromVSCodeSettings_EmptyName(t *testing.T) {
	if cleaned := RemoveContextFromVSCodeSettings(""); cleaned != nil {
		t.Errorf("expected nil for empty context name, got %v", cleaned)
	}
}

func mustJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
