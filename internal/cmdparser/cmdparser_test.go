// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"testing"
)

type TopLevelCommand struct {
	Cmd
}

func (c *TopLevelCommand) DefineCommand(subCommands ...Command) {
	c.Options = Options{
		Use:   "top-level",
		Short: "Hello-World",
		Examples: []ExampleInfo{
			{Description: "First example",
				Steps: []string{"This is the example"}},
		},
	}

	c.Cmd.DefineCommand(subCommands...)
}

type SubCommand1 struct {
	Cmd

	name string
}

func (c *SubCommand1) DefineCommand(subCommands ...Command) {
	c.Options = Options{
		Use:   "sub-command1",
		Short: "Sub Command 1",
		FirstArgAlternativeForFlag: &AlternativeForFlagInfo{
			Flag:  "name",
			Value: &c.name,
		},
		Run: func() { fmt.Println("Running: Sub Command 1") },
	}
	c.Cmd.DefineCommand(subCommands...)
	c.AddFlag(FlagOptions{
		Name:   "name",
		String: &c.name,
		Usage:  "usage",
	})
}

type SubCommand11 struct {
	Cmd
}

func (c *SubCommand11) DefineCommand(...Command) {
	c.Options = Options{
		Use:   "sub-command11",
		Short: "Sub Command 11",
		Run:   func() { fmt.Println("Running: Sub Command 11") },
	}
	c.Cmd.DefineCommand()
}

type SubCommand2 struct {
	Cmd
}

func (c *SubCommand2) DefineCommand(...Command) {
	c.Options = Options{
		Use:     "sub-command2",
		Short:   "Sub Command 2",
		Aliases: []string{"sub-command2-alias"},
	}
	c.Cmd.DefineCommand()
}

func Test_EndToEnd(t *testing.T) {
	subCmd11 := New[*SubCommand11]()
	subCmd1 := New[*SubCommand1](subCmd11)
	subCmd2 := New[*SubCommand2]()

	topLevel := New[*TopLevelCommand](subCmd1, subCmd2)

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

	topLevel.ArgsForUnitTesting([]string{"--help"})
	topLevel.Execute()

	topLevel.ArgsForUnitTesting([]string{"sub-command1", "--help"})
	topLevel.Execute()

	topLevel.ArgsForUnitTesting([]string{"sub-command1", "sub-command11"})
	topLevel.Execute()

	topLevel.ArgsForUnitTesting([]string{"sub-command1"})
	topLevel.Execute()
}

func TestAbstractBase_DefineCommand(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := Cmd{}
	c.DefineCommand()
}
