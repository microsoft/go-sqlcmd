// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmd_run(t *testing.T) {
	s := ""
	c := Cmd{
		options: CommandOptions{
			Use: "test",
			FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
				Flag:  "name",
				Value: &s,
			},
			Examples: []ExampleOptions{{
				Description: "This is an example",
				Steps:       []string{"Step 1", "Step 2"},
			}, {
				Description: "This is a 2nd example",
				Steps:       []string{"Step 1", "Step 2"},
			}},
		},
		dependencies: dependency.Options{Output: output.New(output.Options{ErrorHandler: func(err error) {}, HintHandler: func(hints []string) {}})},
		command:      cobra.Command{},
	}
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.DefineCommand()
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
	assert.Panics(t, func() {
		c.Execute()
	})
}

func TestNegCmdAlternativeValueNotSet(t *testing.T) {
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
	c.SetCrossCuttingConcerns(dependency.Options{
		EndOfLine: "",
		Output:    output.New(output.Options{}),
	})
	c.DefineCommand()
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.SetArgsForUnitTesting([]string{"name-value"})
	assert.Panics(t, func() {
		c.Execute()
	})
}

func TestNegNilOutput(t *testing.T) {
	c := Cmd{
		options: CommandOptions{
			Use: "foo",
			Run: func() {},
		},
	}
	assert.Panics(t, func() {
		c.DefineCommand()
	})
}

func TestNegAddFlag(t *testing.T) {
	TestSetup(t)
	c := Cmd{options: CommandOptions{
		Use: "foo"}}

	assert.Panics(t, func() {
		c.AddFlag(FlagOptions{
			Name:  "",
			Usage: "name",
		})
	})
}

func TestNegAddFlag2(t *testing.T) {
	TestSetup(t)

	c := Cmd{options: CommandOptions{
		Use: "foo"}}

	assert.Panics(t, func() {
		c.AddFlag(FlagOptions{Name: "name", Usage: ""})
	})
}

func TestNegAddFlag3(t *testing.T) {
	TestSetup(t)

	s := "'"
	b := false
	c := Cmd{options: CommandOptions{Use: "foo"}}

	assert.Panics(t, func() {
		c.AddFlag(FlagOptions{Name: "name", Usage: "usage", String: &s, Bool: &b})
	})
}

func TestNegAddFlag4(t *testing.T) {
	TestSetup(t)

	b := false
	i := 0
	c := Cmd{options: CommandOptions{Use: "foo"}}

	assert.Panics(t, func() {
		c.AddFlag(FlagOptions{Name: "name", Usage: "usage", Bool: &b, Int: &i})
	})
}

func TestNegDefineCommandNoCommandOptions(t *testing.T) {
	TestSetup(t)

	c := Cmd{options: CommandOptions{}}

	assert.Panics(t, func() {
		c.DefineCommand()
	})
}

// TestCmd_CheckErrInNotTestingMode covers the code that is not used
// for testing (because we don't want os.Exit() to be called by cobra.checkErr,
// so this test runs in NotTesting mode, and then doesn't pass in an error,
// so the code is covered, but os.Exist isn't called
func TestCmd_CheckErrInNotTestingMode(t *testing.T) {
	TestSetup(t)

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
	TestSetup(t)

	c := Cmd{}

	assert.Panics(t, func() {
		c.Output()
	})
}

func TestNegOutputNewHasNotBeenCalled2(t *testing.T) {
	TestSetup(t)
	c := Cmd{}
	assert.Panics(t, func() {
		c.SetCrossCuttingConcerns(dependency.Options{})
	})
}
