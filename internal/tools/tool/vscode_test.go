// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"reflect"
	"testing"
)

func TestVSCode(t *testing.T) {
	tool := VSCode{}
	tool.Init()

	if tool.Name() != "vscode" {
		t.Errorf("Expected name to be 'vscode', got %s", tool.Name())
	}
}

func TestVSCodeBuildsToSearch(t *testing.T) {
	cases := []struct {
		build string
		want  []string
	}{
		{"", []string{"stable", "insiders"}},
		{"stable", []string{"stable"}},
		{"insiders", []string{"insiders"}},
	}

	for _, c := range cases {
		vscode := VSCode{build: c.build}
		if got := vscode.buildsToSearch(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("build %q: expected %v, got %v", c.build, c.want, got)
		}
	}
}
