// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func (c *Cmd) AddFlag(options FlagOptions) {
	if options.Name == "" {
		panic("Must provide name")
	}
	if options.Usage == "" {
		panic("Must provide usage")
	}

	if options.String != nil {
		if options.Bool != nil || options.Int != nil {
			panic("Only provide one type")
		}
		if options.Shorthand == "" {
			c.command.PersistentFlags().StringVar(
				options.String,
				options.Name,
				options.DefaultString,
				options.Usage)
		} else {
			c.command.PersistentFlags().StringVarP(
				options.String,
				options.Name,
				options.Shorthand,
				options.DefaultString,
				options.Usage)
		}
	}

	if options.Int != nil {
		if options.String != nil || options.Bool != nil {
			panic("Only provide one type")
		}
		if options.Shorthand == "" {
			c.command.PersistentFlags().IntVar(
				options.Int,
				options.Name,
				options.DefaultInt,
				options.Usage)
		} else {
			c.command.PersistentFlags().IntVarP(
				options.Int,
				options.Name,
				options.Shorthand,
				options.DefaultInt,
				options.Usage)
		}
	}

	if options.Bool != nil {
		if options.String != nil || options.Int != nil {
			panic("Only provide one type")
		}
		if options.Shorthand == "" {
			c.command.PersistentFlags().BoolVar(
				options.Bool,
				options.Name,
				options.DefaultBool,
				options.Usage)
		} else {
			c.command.PersistentFlags().BoolVarP(
				options.Bool,
				options.Name,
				options.Shorthand,
				options.DefaultBool,
				options.Usage)
		}
	}
}

func (c *Cmd) ArgsForUnitTesting(args []string) {
	c.command.SetArgs(args)
}

func (c *Cmd) DefineCommand(subCommands ...Command) {
	if c.Options.Use == "" {
		panic("Must implement command definition")
	}

	if c.Options.Long == "" {
		c.Options.Long = c.Options.Short
	}

	c.command = cobra.Command{
		Use:     c.Options.Use,
		Short:   c.Options.Short,
		Long:    c.Options.Long,
		Aliases: c.Options.Aliases,
		Example: c.generateExamples(),
		Run:     c.run,
	}

	if c.Options.FirstArgAlternativeForFlag != nil {
		c.command.Args = cobra.MaximumNArgs(1)
	} else {
		c.command.Args = cobra.MaximumNArgs(0)
	}

	c.addSubCommands(subCommands)
}

// CheckErr passes the error down to cobra.CheckErr (which is likely to call
// os.Exit(1) if err != nil.  Although if running in the golang unit test framework
// we do not want to have os.Exit() called, as this exits the unit test runner
// process, and call panic instead so the call stack can be added to the unit test
// output.
func (c *Cmd) CheckErr(err error) {
	// If we are in a unit test driver, then panic, otherwise pass down to cobra.CheckErr
	if strings.HasSuffix(os.Args[0], ".test") || // are we in go test?
		(len(os.Args) > 1 && os.Args[1] == "-test.v") { // are we in goland unittest?
		if err != nil {
			panic(err)
		}
	} else {
		cobra.CheckErr(err)
	}
}

func (c *Cmd) Command() *cobra.Command {
	return &c.command
}

func (c *Cmd) Execute() {
	err := c.command.Execute()
	c.CheckErr(err)
}

func (c *Cmd) IsSubCommand(command string) (valid bool) {

	if command == "--help" {
		valid = true
	} else if command == "completion" {
		valid = true
	} else {

	outer:
		for _, subCommand := range c.command.Commands() {
			if command == subCommand.Name() {
				valid = true
				break
			}
			for _, alias := range subCommand.Aliases {
				if alias == command {
					valid = true
					break outer
				}
			}
		}
	}
	return
}

func (c *Cmd) addSubCommands(commands []Command) {
	for _, subCommand := range commands {
		c.command.AddCommand(subCommand.Command())
	}
}

func (c *Cmd) generateExamples() string {
	var sb strings.Builder

	for _, e := range c.Options.Examples {
		sb.WriteString(fmt.Sprintf("# %v\n", e.Description))
		for _, s := range e.Steps {
			sb.WriteString(fmt.Sprintf("  %v\n", s))
		}
	}

	return sb.String()
}

func (c *Cmd) run(_ *cobra.Command, args []string) {
	if c.Options.FirstArgAlternativeForFlag != nil {
		if len(args) > 0 {
			flag, err := c.command.PersistentFlags().GetString(
				c.Options.FirstArgAlternativeForFlag.Flag)
			c.CheckErr(err)
			if flag != "" {
				output.Fatal(
					fmt.Sprintf(
						"Both an argument and the --%v flag have been provided. "+
							"Please provide either an argument or the --%v flag",
						c.Options.FirstArgAlternativeForFlag.Flag,
						c.Options.FirstArgAlternativeForFlag.Flag))
			}
			if c.Options.FirstArgAlternativeForFlag.Value == nil {
				panic("Must set Value")
			}
			*c.Options.FirstArgAlternativeForFlag.Value = args[0]
		}
	}

	if c.Options.Run != nil {
		c.Options.Run()
	}
}
