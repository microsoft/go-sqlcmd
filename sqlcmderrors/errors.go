package sqlcmderrors

import (
	"errors"
	"fmt"
)

// ErrorPrefix is the prefix for all sqlcmd-generated errors
const ErrorPrefix = "Sqlcmd: Error: "

// WarningPrefix is the prefix for all sqlcmd-generated warnings
const WarningPrefix = "Sqlcmd: Warning: "

// SQLCmdArgumentError is related to command line switch validation not handled by kong
type SQLCmdArgumentError struct {
	Parameter string
	Rule      string
}

func (e *SQLCmdArgumentError) Error() string {
	return ErrorPrefix + e.Rule
}

// InvalidServerName indicates the SQLCMDSERVER variable has an incorrect format
var InvalidServerName = SQLCmdArgumentError{
	Parameter: "server",
	Rule:      "server must be of the form [tcp]:server[[/instance]|[,port]]",
}

// SQLCmdVariableError is an error about scripting variables
type SQLCmdVariableError struct {
	Variable      string
	MessageFormat string
}

func (e *SQLCmdVariableError) Error() string {
	return ErrorPrefix + fmt.Sprintf(e.MessageFormat, e.Variable)
}

// ReadOnlyVariable indicates the user tried to set a value to a read-only variable
func ReadOnlyVariable(variable string) *SQLCmdVariableError {
	return &SQLCmdVariableError{
		Variable:      variable,
		MessageFormat: "The scripting variable: '%s' is read-only",
	}
}

// SQLCmdCommandError indicates syntax errors for specific sqlcmd commands
type SQLCmdCommandError struct {
	Command    string
	LineNumber uint
}

func (e *SQLCmdCommandError) Error() string {
	return ErrorPrefix + fmt.Sprintf("Syntax error at line %d near command '%s'.", e.LineNumber, e.Command)
}

// InvalidCommandError creates a SQLCmdCommandError
func InvalidCommandError(command string, lineNumber uint) *SQLCmdCommandError {
	return &SQLCmdCommandError{
		Command:    command,
		LineNumber: lineNumber,
	}
}

// InvalidFileError indicates a file could not be opened
func InvalidFileError(err error, path string) error {
	return errors.New(ErrorPrefix + " Error occurred while opening or operating on file " + path + " (Reason: " + err.Error() + ").")
}
