package errors

import (
	"fmt"
)

const errorPrefix = "Sqlcmd: Error: "

// Errors related to command line switches not handled by kong
type SqlCmdArgumentError struct {
	Parameter string
	Rule      string
}

func (e *SqlCmdArgumentError) Error() string {
	return errorPrefix + e.Rule
}

var InvalidServerName = SqlCmdArgumentError{
	Parameter: "server",
	Rule:      "server must be of the form [tcp]:server[[/instance]|[,port]]",
}

// Errors about scripting variables
type SqlCmdVariableError struct {
	Variable      string
	MessageFormat string
}

func (e *SqlCmdVariableError) Error() string {
	return errorPrefix + fmt.Sprintf(e.MessageFormat, e.Variable)
}

func ReadOnlyVariable(variable string) *SqlCmdVariableError {
	return &SqlCmdVariableError{
		Variable:      variable,
		MessageFormat: "The scripting variable: '%s' is read-only",
	}
}

// Syntax errors for specific sqlcmd commands
type SqlCmdCommandError struct {
	Command    string
	LineNumber uint
}

func (e *SqlCmdCommandError) Error() string {
	return errorPrefix + fmt.Sprintf("Syntax error at line %d near command '%s'.", e.LineNumber, e.Command)
}

func InvalidCommandError(command string, lineNumber uint) *SqlCmdCommandError {
	return &SqlCmdCommandError{
		Command:    command,
		LineNumber: lineNumber,
	}
}
