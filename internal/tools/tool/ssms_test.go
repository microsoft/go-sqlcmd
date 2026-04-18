// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import "testing"

func TestSSMS(t *testing.T) {
	tool := SSMS{}
	tool.Init()

	if tool.Name() != "ssms" {
		t.Errorf("Expected name to be 'ssms', got %s", tool.Name())
	}
}
