// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	base := Base{name: "test_tool"}
	assert.Equal(t, base.Name(), "test_tool", "expected 'test_tool', but got %v", base.Name())
}

func TestSetName(t *testing.T) {
	base := Base{}
	base.SetName("test_tool")
	assert.Equal(t, base.Name(), "test_tool", "expected 'test_tool', but got %v", base.Name())
}

func TestExeName(t *testing.T) {
	base := Base{exeName: "test_exe"}
	assert.Equal(t, base.ExeName(), "test_exe", "expected 'test_exe', but got %v", base.ExeName())
}

func TestSetExeName(t *testing.T) {
	base := Base{}
	base.SetExeName("test_exe")
	assert.Equal(t, base.ExeName(), "test_exe", "expected 'test_exe', but got %v", base.ExeName())
}

func TestWhere(t *testing.T) {
	base := Base{exeFullPath: "test/path"}
	assert.Equal(t, base.Where(), "test/path", "expected 'test/path', but got %v", base.Where())
}

func TestIsInstalled(t *testing.T) {
	// Test when exeName is not set
	base := Base{}
	defer func() { assert.NotNil(t, recover(), "The code did not panic as expected") }()
	base.IsInstalled()
}

func TestHowToInstall(t *testing.T) {
	t.Run("windows", func(t *testing.T) {
		b := &Base{
			name: "test",
			description: Description{
				Purpose: "test purpose",
				InstallText: InstallText{
					Windows: "windows install",
					Mac:     "mac install",
					Linux:   "linux install",
				},
			},
		}
		b.HowToInstall()
	})
}
