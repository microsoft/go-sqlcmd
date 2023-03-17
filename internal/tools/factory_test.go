// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tools

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

type TestTool struct {
	tool.Tool
	name string
}

func (t *TestTool) Name() string {
	return t.name
}

func (t *TestTool) Init() {}

func TestNewTool(t *testing.T) {
	// Create some test tools
	tool1 := &TestTool{name: "tool1"}
	tool2 := &TestTool{name: "tool2"}

	// Set up the list of tools
	tools = []tool.Tool{tool1, tool2}

	// Test that we get back the right tool
	result := NewTool("tool1")
	assert.Equal(t, result, tool1, "Expected tool1 but got %v", result)

	// Test that we get an error for an unsupported tool
	expectedErr := fmt.Sprintf("Tool %q is not a supported tool", "unsupported")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic but didn't get one")
		} else if r != expectedErr {
			t.Errorf("Expected panic message %q but got %q", expectedErr, r)
		}
	}()

	_ = NewTool("unsupported")
}
