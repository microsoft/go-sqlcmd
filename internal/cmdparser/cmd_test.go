// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"github.com/spf13/cobra"
	"testing"
)

func TestCmd_run(t *testing.T) {
	s := ""
	c := Cmd{
		options: CommandOptions{
			FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
				Flag:  "name",
				Value: &s,
			},
		},
		dependencies: dependency.Options{Output: output.New(output.Options{ErrorHandler: func(err error) {}, HintHandler: func(hints []string) {}})},
		command:      cobra.Command{},
	}
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.run(nil, []string{"name-value"})
}

func TestNegCmd_run(t *testing.T) {
	s := ""
	c := Cmd{
		options: CommandOptions{
			FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
				Flag:  "name",
				Value: &s,
			}},
		dependencies: dependency.Options{Output: output.New(output.Options{ErrorHandler: func(err error) {}, HintHandler: func(hints []string) {}})},
		command:      cobra.Command{},
	}
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.run(nil, []string{"name-value"})
}

func TestNegCmdProvideBothFlagAndCmd(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	TestSetup(t)
	s := ""
	c := Cmd{
		options: CommandOptions{
			Use: "foo",
			FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
				Flag:  "name",
				Value: &s,
			},
			Run: func() { fmt.Println("Running command") },
		},
		dependencies: dependency.Options{Output: output.New(output.Options{})},
	}
	c.DefineCommand()
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.SetArgsForUnitTesting([]string{"name-value", "--name", "another-value"})
	c.Execute()
}

func TestNegCmdAlternativeValueNotSet(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	s := ""
	c := Cmd{
		options: CommandOptions{
			Use: "foo",
			FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
				Flag:  "name",
				Value: nil,
			},
			Run: func() {},
		},
	}
	c.DefineCommand()
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.SetArgsForUnitTesting([]string{"name-value"})
	c.Execute()
}

func TestNegAddFlag(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	c := Cmd{options: CommandOptions{
		Use: "foo"}}
	c.AddFlag(FlagOptions{
		Name:  "",
		Usage: "name",
	})
}

func TestNegAddFlag2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	c := Cmd{options: CommandOptions{
		Use: "foo"}}
	c.AddFlag(FlagOptions{Name: "name", Usage: ""})
}

func TestNegAddFlag3(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	s := "'"
	b := false
	c := Cmd{options: CommandOptions{Use: "foo"}}
	c.AddFlag(FlagOptions{Name: "name", Usage: "usage", String: &s, Bool: &b})
}

func TestNegAddFlag4(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	b := false
	i := 0
	c := Cmd{options: CommandOptions{Use: "foo"}}
	c.AddFlag(FlagOptions{Name: "name", Usage: "usage", Bool: &b, Int: &i})
}

func TestNegDefineCommandNoCommandOptions(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	c := Cmd{options: CommandOptions{}}
	c.DefineCommand()
}

// TestCmd_CheckErrInNotTestingMode covers the code that is not used
// for testing (because we don't want os.Exit() to be called by cobra.checkErr,
// so this test runs in NotTesting mode, and then doesn't pass in an error,
// so the code is covered, but os.Exist isn't called
func TestCmd_CheckErrInNotTestingMode(t *testing.T) {
	c := Cmd{
		dependencies: dependency.Options{
			EndOfLine: "",
			Output: output.New(output.Options{
				ErrorHandler: func(err error) {},
				HintHandler:  func(hints []string) {},
			}),
		},
		unitTesting: true,
	}
	c.CheckErr(nil)
}

func TestNegOutputNewHasNotBeenCalled(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	c := Cmd{}
	c.Output()
}

func TestNegOutputNewHasNotBeenCalled2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	c := Cmd{}
	c.SetCrossCuttingConcerns(dependency.Options{})
}
