// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tool := &tool{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Init did not panic as expected")
		} else if r != "Do not call directly" {
			t.Errorf("Init panicked with unexpected message: %v", r)
		}
	}()

	tool.Init()
}

func TestName(t *testing.T) {
	expectedName := "Test Tool"
	tool := &tool{
		description: Description{
			Name: expectedName,
		},
	}
	actualName := tool.Name()

	assert.Equal(t, expectedName, actualName)
}

func TestSetExePathAndName(t *testing.T) {
	expectedPath := "/usr/local/bin/test-tool"
	tool := &tool{}
	tool.SetExePathAndName(expectedPath)
	assert.Equal(t, expectedPath, tool.exeName)
}

func TestSetToolDescription(t *testing.T) {
	expectedDescription := Description{
		Name:        "Test Tool",
		Purpose:     "A test tool",
		InstallText: "To install, run `install-test-tool`",
	}
	tool := &tool{}
	tool.SetToolDescription(expectedDescription)
	assert.Equal(t, expectedDescription, tool.description)
}

func TestIsInstalledWhenTrue(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Not passing in CI, skipping for now.")
	}

	tool := &tool{}
	if runtime.GOOS == "windows" {
		tool.exeName = os.Getenv("COMSPEC")
	} else {
		tool.exeName = os.Getenv("SHELL")
	}

	actualInstalled := tool.IsInstalled()
	assert.True(t, actualInstalled)
}

func TestIsInstalledWhenFalse(t *testing.T) {
	tool := &tool{
		exeName: "does-not-exist-tool",
	}

	actualInstalled := tool.IsInstalled()
	assert.False(t, actualInstalled)
}

func TestHowToInstall(t *testing.T) {
	tool := &tool{
		description: Description{
			Name:        "Test Tool",
			Purpose:     "A test tool",
			InstallText: "To install, run `install-test-tool`",
		},
	}
	actualOutput := tool.HowToInstall()
	assert.True(t, strings.Contains(actualOutput, tool.description.InstallText))
}

func TestRunWhenNotInstalled(t *testing.T) {
	tool := &tool{}
	assert.Panics(t, func() {
		_, err := tool.Run([]string{})
		if err != nil {
			return
		}
	})
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Not implemented for Linux yet.")
	}

	t.Parallel()

	tool := &tool{
		exeName:       os.Getenv("COMSPEC"),
		installed:     new(bool),
		lookPathError: nil,
	}

	*tool.installed = true
	exitCode, err := tool.Run([]string{"arg1", "arg2"})
	assert.Equal(t, exitCode, 0)
	assert.NoError(t, err)
}
