// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import "testing"

func TestVSCode(t *testing.T) {
	tool := VSCode{}
	tool.Init()

	if tool.Name() != "vscode" {
		t.Errorf("Expected name to be 'vscode', got %s", tool.Name())
	}
}
