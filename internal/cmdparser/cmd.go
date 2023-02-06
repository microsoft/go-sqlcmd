// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/spf13/cobra"
)

// AddFlag adds a flag to the command instance of type Cmd. The flag is added
// according to the provided FlagOptions. If the FlagOptions does not have a
// name or usage, then the function panics. If the flag is of type String then
// it adds the flag to the PersistentFlags of the command instance with the
// provided options. Similarly, if the flag is of type Int or Bool, the flag
// is added to the PersistentFlags with the provided options. If a shorthand
// is provided, then it uses StringVarP or BoolVarP or IntVarP instead of
// StringVar or BoolVar or IntVar respectively.
func (c *Cmd) AddFlag(options FlagOptions) {
	if options.Name == "" {
		panic("Must provide name")
	}
	if options.Usage == "" {
		panic("Must provide usage for flag")
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
		if options.Bool != nil {
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

// DefineCommand defines a command with the provided CommandOptions and adds
// it to the command list. If only one CommandOptions is provided, it is used
// as the command options. Otherwise, the default CommandOptions are used. The
// function sets the command usage, short and long descriptions, aliases, examples,
// and run function. It also sets the maximum number of arguments allowed for the
// command, and adds any subcommands specified in the CommandOptions.
func (c *Cmd) DefineCommand(options ...CommandOptions) {
	if len(options) == 1 {
		c.options = options[0]
	}

	if c.options.Use == "" {
		panic("Must implement command definition")
	}

	if c.options.Long == "" {
		c.options.Long = c.options.Short
	}

	c.command = cobra.Command{
		Use:     c.options.Use,
		Short:   c.options.Short,
		Long:    c.options.Long,
		Aliases: c.options.Aliases,
		Example: c.generateExamples(),
		Run:     c.run,
	}

	if c.options.FirstArgAlternativeForFlag != nil {
		c.command.Args = cobra.MaximumNArgs(1)

		// IDIOMATIC: override the Use so the --help includes the flag name in caps and square bracket
		// e.g. `sqlcmd config use-context [NAME]` or `sqlcmd config delete-user [NAME]`
		c.command.Use = c.options.Use + " [" + strings.ToUpper(c.options.FirstArgAlternativeForFlag.Flag) + "]"
	} else {
		c.command.Args = cobra.MaximumNArgs(0)
	}

	c.addSubCommands(c.options.SubCommands)
}

// CheckErr passes the error down to cobra.CheckErr (which is likely to call
// os.Exit(1) if err != nil.  Although if running in the golang unit test framework
// we do not want to have os.Exit() called, as this exits the unit test runner
// process, and call panic instead so the call stack can be added to the unit test
// output.
func (c *Cmd) CheckErr(err error) {
	output := c.Output()
	output.FatalErr(err)
}

// Command returns the cobra Command associated with the Cmd. This method
// allows for easy access and manipulation of the command's properties and behavior.
func (c *Cmd) Command() *cobra.Command {
	return &c.command
}

// Execute function is responsible for executing the underlying command for
// this Cmd object. The function first attempts to execute the command, and then
// checks for any errors that may have occurred during execution. If an error
// is detected, the CheckErr method is called to handle the error. This function
// is typically called after defining and configuring the command using
// the DefineCommand and SetArgsForUnitTesting functions.
func (c *Cmd) Execute() {
	err := c.command.Execute()
	c.CheckErr(err)
}

// Output function is a getter function that returns the output.Output instance
// associated with the Cmd instance. If no output.Output instance has been
// set, the function initializes a new instance and returns it.
func (c *Cmd) Output() *output.Output {
	if c.dependencies.Output == nil {
		panic("output.New has not been called yet (call SetCrossCuttingConcerns first?)")
	}
	return c.dependencies.Output
}

func (c *Cmd) Dependencies() dependency.Options {
	return c.dependencies
}

// Inject dependencies into the Cmd struct. The options parameter is a struct
// containing a reference to the output struct, which the function then
// assigns to the output field of the Cmd struct. This allows for the
// output struct to be mocked in unit tests.
func (c *Cmd) SetCrossCuttingConcerns(dependencies dependency.Options) {
	if dependencies.Output == nil {
		panic("Output is nil")
	}
	c.dependencies = dependencies
}

// IsSubCommand returns true if the provided command string
// matches the name or an alias of one of the object sub-commands,
// or if the command string is "--help" or "completion". Otherwise,
// it returns false.
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

// SetArgsForUnitTesting sets the arguments for a unit test.
// This function allows users to specify arguments to the command for testing purposes.
func (c *Cmd) SetArgsForUnitTesting(args []string) {
	c.command.SetArgs(args)
}

// addSubCommands is a helper function that is used to add multiple sub-commands
// to a parent command in the application. It takes a slice of Command objects
// as an input and then adds each Command object to the parent command using
// the AddCommand method. This allows for a modular approach to defining
// the application's command hierarchy and makes it easy to add new sub-commands
// to the parent command.
func (c *Cmd) addSubCommands(commands []Command) {
	if c.dependencies.Output == nil {
		panic("Why is output nil?")
	}
	for _, subCommand := range commands {
		c.command.AddCommand(subCommand.Command())
	}
}

// generateExamples generates a list of examples for a command. It iterates
// over the Examples property of the CommandOptions struct, appending the
// description and steps for each example. The resulting string is returned.
func (c *Cmd) generateExamples() string {
	var sb strings.Builder

	for i, e := range c.options.Examples {
		sb.WriteString(fmt.Sprintf("# %v%v", e.Description, pal.LineBreak()))
		for ii, s := range e.Steps {
			sb.WriteString(fmt.Sprintf("  - %v", s))
			if ii != len(e.Steps)-1 {
				sb.WriteString(pal.LineBreak())
			}
		}
		if i != len(c.options.Examples)-1 {
			sb.WriteString(pal.LineBreak())
		}
	}

	return sb.String()
}

// run function is a command handler for the cobra library. It checks if the first
// argument has been provided as an alternative for the specified flag and, if so,
// sets the value of that flag to the provided argument. If the Run option has been
// specified in the CommandOptions, it calls that function.
func (c *Cmd) run(_ *cobra.Command, args []string) {
	if c.options.FirstArgAlternativeForFlag != nil {
		if len(args) > 0 {
			flag, err := c.command.PersistentFlags().GetString(
				c.options.FirstArgAlternativeForFlag.Flag)
			c.CheckErr(err)

			if flag != "" {
				c.dependencies.Output.Fatal(
					fmt.Sprintf(
						"Both an argument and the --%v flag have been provided. "+
							"Please provide either an argument or the --%v flag",
						c.options.FirstArgAlternativeForFlag.Flag,
						c.options.FirstArgAlternativeForFlag.Flag))
			}
			if c.options.FirstArgAlternativeForFlag.Value == nil {
				panic("Must set Value")
			}
			*c.options.FirstArgAlternativeForFlag.Value = args[0]
		}
	}

	if c.options.Run == nil {
		// If command has no run, it has sub-commands only, then display help if no
		// sub-command entered
		err := c.command.Help()
		c.CheckErr(err)
	} else {
		c.options.Run()
	}
}
