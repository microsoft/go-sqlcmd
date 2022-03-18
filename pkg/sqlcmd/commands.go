// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
)

// Command defines a sqlcmd action which can be intermixed with the SQL batch
// Commands for sqlcmd are defined at https://docs.microsoft.com/sql/tools/sqlcmd-utility#sqlcmd-commands
type Command struct {
	// regex must include at least one group if it has parameters
	// Will be matched using FindStringSubmatch
	regex *regexp.Regexp
	// The function that implements the command. Third parameter is the line number
	action func(*Sqlcmd, []string, uint) error
	// Name of the command
	name string
}

// Commands is the set of sqlcmd command implementations
type Commands map[string]*Command

func newCommands() Commands {
	// Commands is the set of Command implementations
	return map[string]*Command{
		"EXIT": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:?EXIT(?:[ \t]*(\(?.*\)?$)|$)`),
			action: exitCommand,
			name:   "EXIT",
		},
		"QUIT": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:?QUIT(?:[ \t]+(.*$)|$)`),
			action: quitCommand,
			name:   "QUIT",
		},
		"GO": {
			regex:  regexp.MustCompile(batchTerminatorRegex("GO")),
			action: goCommand,
			name:   "GO",
		},
		"OUT": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:OUT(?:[ \t]+(.*$)|$)`),
			action: outCommand,
			name:   "OUT",
		},
		"ERROR": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:ERROR(?:[ \t]+(.*$)|$)`),
			action: errorCommand,
			name:   "ERROR",
		},
		"READFILE": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:R(?:[ \t]+(.*$)|$)`),
			action: readFileCommand,
			name:   "READFILE",
		},
		"SETVAR": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:SETVAR(?:[ \t]+(.*$)|$)`),
			action: setVarCommand,
			name:   "SETVAR",
		},
		"LISTVAR": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:LISTVAR(?:[ \t]+(.*$)|$)`),
			action: listVarCommand,
			name:   "LISTVAR",
		},
		"RESET": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:RESET(?:[ \t]+(.*$)|$)`),
			action: resetCommand,
			name:   "RESET",
		},
		"LIST": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:LIST(?:[ \t]+(.*$)|$)`),
			action: listCommand,
			name:   "LIST",
		},
		"CONNECT": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:CONNECT(?:[ \t]+(.*$)|$)`),
			action: connectCommand,
			name:   "CONNECT",
		},
	}

}

func (c Commands) matchCommand(line string) (*Command, []string) {
	for _, cmd := range c {
		matchedCommand := cmd.regex.FindStringSubmatch(line)
		if matchedCommand != nil {
			return cmd, matchedCommand[1:]
		}
	}
	return nil, nil
}

func batchTerminatorRegex(terminator string) string {
	return fmt.Sprintf(`(?im)^[\t ]*?%s(?:[ ]+(.*$)|$)`, regexp.QuoteMeta(terminator))
}

// SetBatchTerminator attempts to set the batch terminator to the given value
// Returns an error if the new value is not usable in the regex
func (c Commands) SetBatchTerminator(terminator string) error {
	cmd := c["GO"]
	regex, err := regexp.Compile(batchTerminatorRegex(terminator))
	if err != nil {
		return err
	}
	cmd.regex = regex
	return nil
}

// exitCommand has 3 modes.
// With no (), it just exits without running any query
// With () it runs whatever batch is in the buffer then exits
// With any text between () it runs the text as a query then exits
func exitCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 {
		return ErrExitRequested
	}
	params := strings.TrimSpace(args[0])
	if params == "" {
		return ErrExitRequested
	}
	if !strings.HasPrefix(params, "(") || !strings.HasSuffix(params, ")") {
		return InvalidCommandError("EXIT", line)
	}
	// First we run the current batch
	query := s.batch.String()
	if query != "" {
		query = s.getRunnableQuery(query)
		if exitCode, err := s.runQuery(query); err != nil {
			s.Exitcode = exitCode
			return ErrExitRequested
		}
	}
	query = strings.TrimSpace(params[1 : len(params)-1])
	if query != "" {
		query = s.getRunnableQuery(query)
		s.Exitcode, _ = s.runQuery(query)
	}
	return ErrExitRequested
}

// quitCommand immediately exits the program without running any more batches
func quitCommand(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return InvalidCommandError("QUIT", line)
	}
	return ErrExitRequested
}

// goCommand runs the current batch the number of times specified
func goCommand(s *Sqlcmd, args []string, line uint) error {
	// default to 1 execution
	n := 1
	var err error
	if len(args) > 0 {
		cnt := strings.TrimSpace(args[0])
		if cnt != "" {
			_, err = fmt.Sscanf(cnt, "%d", &n)
		}
	}
	if err != nil || n < 1 {
		return InvalidCommandError("GO", line)
	}
	query := s.batch.String()
	if query == "" {
		return nil
	}
	query = s.getRunnableQuery(query)
	for i := 0; i < n; i++ {
		if retcode, err := s.runQuery(query); err != nil {
			s.Exitcode = retcode
			return err
		}
	}
	s.batch.Reset(nil)
	return nil
}

// outCommand changes the output writer to use a file
func outCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 || args[0] == "" {
		return InvalidCommandError("OUT", line)
	}
	switch {
	case strings.EqualFold(args[0], "stdout"):
		s.SetOutput(os.Stdout)
	case strings.EqualFold(args[0], "stderr"):
		s.SetOutput(os.Stderr)
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return InvalidFileError(err, args[0])
		}
		s.SetOutput(o)
	}
	return nil
}

// errorCommand changes the error writer to use a file
func errorCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 || args[0] == "" {
		return InvalidCommandError("OUT", line)
	}
	switch {
	case strings.EqualFold(args[0], "stderr"):
		s.SetError(os.Stderr)
	case strings.EqualFold(args[0], "stdout"):
		s.SetError(os.Stdout)
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return InvalidFileError(err, args[0])
		}
		s.SetError(o)
	}
	return nil
}

func readFileCommand(s *Sqlcmd, args []string, line uint) error {
	if args == nil || len(args) != 1 {
		return InvalidCommandError(":R", line)
	}
	return s.IncludeFile(args[0], false)
}

// setVarCommand parses a variable setting and applies it to the current Sqlcmd variables
func setVarCommand(s *Sqlcmd, args []string, line uint) error {
	if args == nil || len(args) != 1 || args[0] == "" {
		return InvalidCommandError(":SETVAR", line)
	}

	varname := args[0]
	val := ""
	// The prior incarnation of sqlcmd doesn't require a space between the variable name and its value
	// in some very unexpected cases. This version will require the space.
	sp := strings.IndexRune(args[0], ' ')
	if sp > -1 {
		val = strings.TrimSpace(varname[sp:])
		varname = varname[:sp]
	}
	if err := s.vars.Setvar(varname, val); err != nil {
		switch e := err.(type) {
		case *VariableError:
			return e
		default:
			return InvalidCommandError(":SETVAR", line)
		}
	}
	return nil
}

// listVarCommand prints the set of Sqlcmd scripting variables.
// Builtin values are printed first, followed by user-set values in sorted order.
func listVarCommand(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return InvalidCommandError("LISTVAR", line)
	}

	vars := s.vars.All()
	keys := make([]string, 0, len(vars))
	for k := range vars {
		if !contains(builtinVariables, k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	keys = append(builtinVariables, keys...)
	for _, k := range keys {
		fmt.Fprintf(s.GetOutput(), `%s = "%s"%s`, k, vars[k], SqlcmdEol)
	}
	return nil
}

// resetCommand resets the statement cache
func resetCommand(s *Sqlcmd, args []string, line uint) error {
	if s.batch != nil {
		s.batch.Reset(nil)
	}

	return nil
}

// listCommand displays statements currently in  the statement cache
func listCommand(s *Sqlcmd, args []string, line uint) error {
	if s.batch != nil && s.batch.String() != "" {
		fmt.Fprintf(s.GetOutput(), `%s%s`, []byte(s.batch.String()), SqlcmdEol)
	}

	return nil
}

type connectData struct {
	Server               string `arg:""`
	Database             string `short:"D"`
	Username             string `short:"U"`
	Password             string `short:"P"`
	LoginTimeout         int    `short:"l"`
	AuthenticationMethod string `short:"G"`
}

func connectCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return InvalidCommandError("CONNECT", line)
	}
	arguments := &connectData{}
	parser, err := kong.New(arguments)
	if err != nil {
		return InvalidCommandError("CONNECT", line)
	}
	if _, err = parser.Parse(strings.Split(args[0], " ")); err != nil {
		return InvalidCommandError("CONNECT", line)
	}

	connect := s.Connect
	connect.UserName = arguments.Username
	connect.Password = arguments.Password
	connect.ServerName = arguments.Server
	if arguments.LoginTimeout > 0 {
		connect.LoginTimeoutSeconds = arguments.LoginTimeout
	}
	connect.AuthenticationMethod = arguments.AuthenticationMethod
	// If no user name is provided we switch to integrated auth
	_ = s.ConnectDb(&connect, s.lineIo == nil)
	// ConnectDb prints connection errors already, and failure to connect is not fatal even with -b option
	return nil
}
