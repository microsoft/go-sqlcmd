// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"errors"
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

type TopLevelCommand struct {
	Cmd
}

func (c *TopLevelCommand) DefineCommand(...CommandOptions) {
	commandOptions := CommandOptions{
		Use:   "top-level",
		Short: "Hello-World",
		Examples: []ExampleOptions{
			{Description: "First example",
				Steps: []string{"This is the example"}},
		},
		SubCommands: c.SubCommands(),
	}

	c.Cmd.DefineCommand(commandOptions)
}

func (c *TopLevelCommand) SubCommands() []Command {
	return []Command{
		New[*SubCommand1](c.Dependencies()),
		New[*SubCommand2](c.Dependencies()),
		New[*ErrorCommand](c.Dependencies()),
	}
}

type SubCommand1 struct {
	Cmd

	name string
}

func (c *SubCommand1) DefineCommand(...CommandOptions) {
	commandOptions := CommandOptions{
		Use:   "sub-command1",
		Short: "Sub Command 1",
		FirstArgAlternativeForFlag: &AlternativeForFlagOptions{
			Flag:  "name",
			Value: &c.name,
		},
		Run: func() {
			c.Output().InfofWithHints([]string{"This is a hint"}, "This is a message")
		},
		SubCommands: c.SubCommands(),
	}
	c.Cmd.DefineCommand(commandOptions)
	c.AddFlag(FlagOptions{
		Name:   "name",
		String: &c.name,
		Usage:  "usage",
	})
}

func (c *SubCommand1) SubCommands() []Command {
	return []Command{
		New[*SubCommand11](c.Dependencies()),
	}
}

type SubCommand11 struct {
	Cmd
}

func (c *SubCommand11) DefineCommand(...CommandOptions) {
	commandOptions := CommandOptions{
		Use:   "sub-command11",
		Short: "Sub Command 11",
		Run:   func() { fmt.Println("Running: Sub Command 11") },
	}
	c.Cmd.DefineCommand(commandOptions)
}

type SubCommand2 struct {
	Cmd
}

func (c *SubCommand2) DefineCommand(...CommandOptions) {
	commandOptions := CommandOptions{
		Use:     "sub-command2",
		Short:   "Sub Command 2",
		Aliases: []string{"sub-command2-alias"},
	}
	c.Cmd.DefineCommand(commandOptions)
}

type ErrorCommand struct {
	Cmd
}

func (c *ErrorCommand) DefineCommand(...CommandOptions) {
	commandOptions := CommandOptions{
		Use:   "error-command",
		Short: "Generate an error",
		Run:   c.run,
	}
	c.Cmd.DefineCommand(commandOptions)
}

func (c *ErrorCommand) run() {
	output := c.dependencies.Output

	output.Fatal("This command causes the cli to exit")
}

func Test_EndToEnd(t *testing.T) {
	topLevel := New[*TopLevelCommand](dependency.Options{})

	topLevel.IsSubCommand("sub-command2")
	topLevel.IsSubCommand("sub-command2-alias")
	topLevel.IsSubCommand("--help")
	topLevel.IsSubCommand("completion")

	var s string
	topLevel.AddFlag(FlagOptions{
		String: &s,
		Name:   "string",
		Usage:  "usage",
	})
	topLevel.AddFlag(FlagOptions{
		String:    &s,
		Shorthand: "s",
		Name:      "string2",
		Usage:     "usage",
	})

	var i int
	topLevel.AddFlag(FlagOptions{
		Int:   &i,
		Name:  "int",
		Usage: "usage",
	})
	topLevel.AddFlag(FlagOptions{
		Int:       &i,
		Shorthand: "i",
		Name:      "int2",
		Usage:     "usage",
	})

	var b bool
	topLevel.AddFlag(FlagOptions{
		Bool:  &b,
		Name:  "bool",
		Usage: "usage",
	})
	topLevel.AddFlag(FlagOptions{
		Bool:      &b,
		Shorthand: "b",
		Name:      "bool2",
		Usage:     "usage",
	})

	topLevel.SetArgsForUnitTesting([]string{"--help"})
	topLevel.Execute()

	topLevel.SetArgsForUnitTesting([]string{"sub-command1", "--help"})
	topLevel.Execute()

	topLevel.SetArgsForUnitTesting([]string{"sub-command1", "sub-command11"})
	topLevel.Execute()

	topLevel.SetArgsForUnitTesting([]string{"sub-command1"})
	topLevel.Execute()
}

func TestInitialize(t *testing.T) {
	Initialize(func() { fmt.Println("Got here") })
}

func Test(t *testing.T) {
	topLevel := New[*TopLevelCommand](dependency.Options{})
	topLevel.SetArgsForUnitTesting([]string{})
	topLevel.CheckErr(nil)

	topLevel = New[*TopLevelCommand](dependency.Options{})
	topLevel.SetArgsForUnitTesting([]string{})
	topLevel.CheckErr(nil)
}

func Test2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	topLevel := New[*TopLevelCommand](dependency.Options{})
	topLevel.SetArgsForUnitTesting([]string{})
	topLevel.CheckErr(errors.New("foo"))
}
