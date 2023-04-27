// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

// The AlternativeForFlagOptions type represents options for defining an alternative
// for a flag. It consists of the name of the flag, as well as a pointer to the
// value to be used as the alternative. This type is typically used in the case
// where the user has provided an argument that should be treated as an alternative
// to a specific flag.
type AlternativeForFlagOptions struct {
	Flag  string
	Value *string
}

// FlagOptions type represents options for defining a flag for a command-line
// interface. The Name and Shorthand fields specify the long and short names
// for the flag, respectively. The Usage field is a string that describes how the
// flag should be used. The String, DefaultString, Int, DefaultInt, Bool, and
// DefaultBool fields are used to specify the type and default value of the flag,
// if it is a string, int, or bool type. The String and Int fields should be pointers
// to the variables that will store the flag's value, and the Bool field should be
// a pointer to a bool variable that will be set to true if the flag is present. The
// DefaultString, DefaultInt, and DefaultBool fields are the default values to
// use if the flag is not provided by the user.
type FlagOptions struct {
	Name      string
	Shorthand string
	Usage     string

	Hidden bool

	String        *string
	DefaultString string

	Int        *int
	DefaultInt int

	Bool        *bool
	DefaultBool bool
}

// CommandOptions is a struct that allows the caller to specify options for a Command.
// These options include the command's name, description, usage, and behavior.
// The Aliases field specifies alternate names for the command,
// and the Examples field specifies examples of how to use the command.
// The FirstArgAlternativeForFlag field specifies an alternative to the first
// argument when it is provided as a flag, and the Long and Short fields
// specify the command's long and short descriptions, respectively.
// The Run field specifies the behavior of the command when it is executed,
// and the Use field specifies the usage instructions for the command.
// The SubCommands field specifies any subcommands that the command has.
type CommandOptions struct {
	Aliases                    []string
	Examples                   []ExampleOptions
	FirstArgAlternativeForFlag *AlternativeForFlagOptions
	Long                       string
	Run                        func()
	Short                      string
	Use                        string
	SubCommands                []Command
}

// ExampleOptions specifies the details of an example usage of a command.
// It contains a description of the example, and a list of steps that make up
// the example. This type is typically used in conjunction with the Examples
// field of the CommandOptions struct, to provide examples of how to use a
// command in the command's help text.
type ExampleOptions struct {
	Description string
	Steps       []string
}
