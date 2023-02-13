// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilenameInUserHomeDotDirectory(t *testing.T) {
	FilenameInUserHomeDotDirectory(".foo", "bar")
}

func TestLineBreak(t *testing.T) {
	LineBreak()
}

func TestNegLineBreak(t *testing.T) {
	assert.Panics(t, func() {
		lineBreak = ""
		LineBreak()
	})
}

func TestCheckErr(t *testing.T) {
	assert.Panics(t, func() {
		checkErr(errors.New("test"))
	})
}

func TestUserName(t *testing.T) {
	user := UserName()
	fmt.Println(user)
}

func TestCmdLineWithEnvVars(t *testing.T) {
	cmdLine := CmdLineWithEnvVars([]string{"ENVVAR=FOOBAR"}, "cmd-to-run.exe")
	fmt.Println(cmdLine)
}
