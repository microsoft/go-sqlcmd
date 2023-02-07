// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorPrefix is the prefix for all sqlcmd-generated errors
const ErrorPrefix = "Sqlcmd: Error: "

// WarningPrefix is the prefix for all sqlcmd-generated warnings
const WarningPrefix = "Sqlcmd: Warning: "

// Common Sqlcmd error messages
const ErrCmdDisabled = "ED and !!<command> commands, startup script, and environment variables are disabled"

type SqlcmdError interface {
	error
	IsSqlcmdErr() bool
}

type CommonSqlcmdErr struct {
	message string
}

func (e *CommonSqlcmdErr) Error() string {
	return e.message
}

func (e *CommonSqlcmdErr) IsSqlcmdErr() bool {
	return true
}

// ArgumentError is related to command line switch validation not handled by kong
type ArgumentError struct {
	Parameter string
	Rule      string
}

func (e *ArgumentError) Error() string {
	return ErrorPrefix + e.Rule
}

func (e *ArgumentError) IsSqlcmdErr() bool {
	return true
}

// InvalidServerName indicates the SQLCMDSERVER variable has an incorrect format
var InvalidServerName = ArgumentError{
	Parameter: "server",
	Rule:      "server must be of the form [[np]|[lpc][tcp]]:server[[/instance]|[,port]]",
}

// VariableError is an error about scripting variables
type VariableError struct {
	Variable      string
	MessageFormat string
}

func (e *VariableError) Error() string {
	return ErrorPrefix + fmt.Sprintf(e.MessageFormat, e.Variable)
}

func (e *VariableError) IsSqlcmdErr() bool {
	return true
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

// InvalidVariableValue indicates the variable was set to an invalid value
func InvalidVariableValue(variable string, value string) *VariableError {
	return &VariableError{
		Variable:      variable,
		MessageFormat: "The environment variable: '%s' has invalid value: '" + strings.ReplaceAll(value, `%`, `%%`) + "'.",
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

func (e *CommandError) IsSqlcmdErr() bool {
	return true
}

// InvalidCommandError creates a SQLCmdCommandError
func InvalidCommandError(command string, lineNumber uint) *CommandError {
	return &CommandError{
		Command:    command,
		LineNumber: lineNumber,
	}
}

type FileError struct {
	err  error
	path string
}

func (e *FileError) Error() string {
	return e.err.Error()
}

func (e *FileError) IsSqlcmdErr() bool {
	return true
}

// InvalidFileError indicates a file could not be opened
func InvalidFileError(err error, filepath string) error {
	return &FileError{
		err:  errors.New(ErrorPrefix + " Error occurred while opening or operating on file " + filepath + " (Reason: " + err.Error() + ")."),
		path: filepath,
	}
}

type SyntaxError struct {
	err error
}

func (e *SyntaxError) Error() string {
	return e.err.Error()
}

func (e *SyntaxError) IsSqlcmdErr() bool {
	return true
}

// SyntaxError indicates a malformed sqlcmd statement
func syntaxError(lineNumber uint) SqlcmdError {
	return &SyntaxError{
		err: fmt.Errorf("%sSyntax error at line %d", ErrorPrefix, lineNumber),
	}
}
