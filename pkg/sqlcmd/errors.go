// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"errors"
	"fmt"
)

// ErrorPrefix is the prefix for all sqlcmd-generated errors
const ErrorPrefix = "Sqlcmd: Error: "

// WarningPrefix is the prefix for all sqlcmd-generated warnings
const WarningPrefix = "Sqlcmd: Warning: "

// ArgumentError is related to command line switch validation not handled by kong
type ArgumentError struct {
	Parameter string
	Rule      string
}

func (e *ArgumentError) Error() string {
	return ErrorPrefix + e.Rule
}

// InvalidServerName indicates the SQLCMDSERVER variable has an incorrect format
var InvalidServerName = ArgumentError{
	Parameter: "server",
	Rule:      "server must be of the form [tcp]:server[[/instance]|[,port]]",
}

// VariableError is an error about scripting variables
type VariableError struct {
	Variable      string
	MessageFormat string
}

func (e *VariableError) Error() string {
	return ErrorPrefix + fmt.Sprintf(e.MessageFormat, e.Variable)
}

// ReadOnlyVariable indicates the user tried to set a value to a read-only variable
func ReadOnlyVariable(variable string) *VariableError {
	return &VariableError{
		Variable:      variable,
		MessageFormat: "The scripting variable: '%s' is read-only",
	}
}

// UndefinedVariable indicates the user tried to reference an undefined variable
func UndefinedVariable(variable string) *VariableError {
	return &VariableError{
		Variable:      variable,
		MessageFormat: "'%s' scripting variable not defined.",
	}
}

// CommandError indicates syntax errors for specific sqlcmd commands
type CommandError struct {
	Command    string
	LineNumber uint
}

func (e *CommandError) Error() string {
	return ErrorPrefix + fmt.Sprintf("Syntax error at line %d near command '%s'.", e.LineNumber, e.Command)
}

// InvalidCommandError creates a SQLCmdCommandError
func InvalidCommandError(command string, lineNumber uint) *CommandError {
	return &CommandError{
		Command:    command,
		LineNumber: lineNumber,
	}
}

// InvalidFileError indicates a file could not be opened
func InvalidFileError(err error, path string) error {
	return errors.New(ErrorPrefix + " Error occurred while opening or operating on file " + path + " (Reason: " + err.Error() + ").")
}

// SyntaxError indicates a malformed sqlcmd statement
func syntaxError(lineNumber uint) error {
	return fmt.Errorf("%sSyntax error at line %d.", ErrorPrefix, lineNumber)
}
