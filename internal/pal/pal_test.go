// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"errors"
	"fmt"
	"testing"
)

func TestFilenameInUserHomeDotDirectory(t *testing.T) {
	FilenameInUserHomeDotDirectory(".foo", "bar")
}

func TestCheckErr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	checkErr(errors.New("test"))
}

func TestUserName(t *testing.T) {
	user := UserName()
	fmt.Println(user)
}

func TestCmdLineWithEnvVars(t *testing.T) {
	cmdLine := CmdLineWithEnvVars([]string{"ENVVAR=FOOBAR"}, "cmd-to-run.exe")
	fmt.Println(cmdLine)
}
