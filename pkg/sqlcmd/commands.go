// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/color"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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
	// whether the command is a system command
	isSystem bool
}

// Commands is the set of sqlcmd command implementations
type Commands map[string]*Command

func newCommands() Commands {
	// Commands is the set of Command implementations
	return map[string]*Command{
		"EXIT": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:?EXIT([\( \t]+.*\)*$|$)`),
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
		}, "READFILE": {
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
			regex:  regexp.MustCompile(`(?im)^[ \t]*?:?RESET(?:[ \t]+(.*$)|$)`),
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
		"EXEC": {
			regex:    regexp.MustCompile(`(?im)^[ \t]*?:?!!(.*$)`),
			action:   execCommand,
			name:     "EXEC",
			isSystem: true,
		},
		"EDIT": {
			regex:    regexp.MustCompile(`(?im)^[\t ]*?:?ED(?:[ \t]+(.*$)|$)`),
			action:   editCommand,
			name:     "EDIT",
			isSystem: true,
		},
		"ONERROR": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:?ON ERROR(?:[ \t]+(.*$)|$)`),
			action: onerrorCommand,
			name:   "ONERROR",
		},
		"XML": {
			regex:  regexp.MustCompile(`(?im)^[\t ]*?:XML(?:[ \t]+(.*$)|$)`),
			action: xmlCommand,
			name:   "XML",
		},
		"HELP": {
			regex:  regexp.MustCompile(`(?im)^[ \t]*:HELP(?:[ \t]+(.*$)|$)`),
			action: helpCommand,
			name:   "HELP",
		},
	}
}

// DisableSysCommands disables the ED and :!! commands.
// When exitOnCall is true, running those commands will exit the process.
func (c Commands) DisableSysCommands(exitOnCall bool) {
	f := warnDisabled
	if exitOnCall {
		f = errorDisabled
	}
	for _, cmd := range c {
		if cmd.isSystem {
			cmd.action = f
		}
	}
}

func (c Commands) matchCommand(line string) (*Command, []string) {
	for _, cmd := range c {
		matchedCommand := cmd.regex.FindStringSubmatch(line)
		if matchedCommand != nil {
			return cmd, removeComments(matchedCommand[1:])
		}
	}
	return nil, nil
}

func removeComments(args []string) []string {
	var pos int
	quote := false
	for i := range args {
		pos, quote = commentStart([]rune(args[i]), quote)
		if pos > -1 {
			out := make([]string, i+1)
			if i > 0 {
				copy(out, args[:i])
			}
			out[i] = args[i][:pos]
			return out
		}
	}
	return args
}

func commentStart(arg []rune, quote bool) (int, bool) {
	var i int
	space := true
	for ; i < len(arg); i++ {
		c, next := arg[i], grab(arg, i+1, len(arg))
		switch {
		case quote && c == '"' && next != '"':
			quote = false
		case quote && c == '"' && next == '"':
			i++
		case c == '\t' || c == ' ':
			space = true
		// Note we assume none of the regexes would split arguments on non-whitespace boundaries such that "text -- comment" would get split into "text -" and "- comment"
		case !quote && space && c == '-' && next == '-':
			return i, false
		case !quote && c == '"':
			quote = true
		default:
			space = false
		}
	}
	return -1, quote
}

func warnDisabled(s *Sqlcmd, args []string, line uint) error {
	s.WriteError(s.GetError(), ErrCommandsDisabled)
	return nil
}

func errorDisabled(s *Sqlcmd, args []string, line uint) error {
	s.WriteError(s.GetError(), ErrCommandsDisabled)
	s.Exitcode = 1
	return ErrExitRequested
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
	// First we save the current batch
	query1 := s.batch.String()
	if len(query1) > 0 {
		query1 = s.getRunnableQuery(query1)
	}
	// Now parse the params of EXIT as a batch without commands
	cmd := s.batch.cmd
	s.batch.cmd = nil
	defer func() {
		s.batch.cmd = cmd
	}()
	query2 := strings.TrimSpace(params[1 : len(params)-1])
	if len(query2) > 0 {
		s.batch.Reset([]rune(query2))
		_, _, err := s.batch.Next()
		if err != nil {
			return err
		}
		query2 = s.batch.String()
		if len(query2) > 0 {
			query2 = s.getRunnableQuery(query2)
		}
	}

	if len(query1) > 0 || len(query2) > 0 {
		query := query1 + SqlcmdEol + query2
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
			if cnt, err = resolveArgumentVariables(s, []rune(cnt), true); err != nil {
				return err
			}
			_, err = fmt.Sscanf(cnt, "%d", &n)
		}
	}
	if err != nil || n < 1 {
		return InvalidCommandError("GO", line)
	}
	if s.EchoInput {
		err = listCommand(s, []string{}, line)
	}
	if err != nil {
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
	filePath, err := resolveArgumentVariables(s, []rune(args[0]), true)
	if err != nil {
		return err
	}

	switch {
	case strings.EqualFold(filePath, "stdout"):
		s.SetOutput(os.Stdout)
	case strings.EqualFold(filePath, "stderr"):
		s.SetOutput(os.Stderr)
	default:
		o, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return InvalidFileError(err, args[0])
		}
		if s.UnicodeOutputFile {
			// ODBC sqlcmd doesn't write a BOM but we will.
			// Maybe the endian-ness should be configurable.
			win16le := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
			encoder := transform.NewWriter(o, win16le.NewEncoder())
			s.SetOutput(encoder)
		} else {
			s.SetOutput(o)
		}
	}
	return nil
}

// errorCommand changes the error writer to use a file
func errorCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 || args[0] == "" {
		return InvalidCommandError("ERROR", line)
	}
	filePath, err := resolveArgumentVariables(s, []rune(args[0]), true)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(filePath, "stderr"):
		s.SetError(os.Stderr)
	case strings.EqualFold(filePath, "stdout"):
		s.SetError(os.Stdout)
	default:
		o, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
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
	fileName, _ := resolveArgumentVariables(s, []rune(args[0]), false)
	return s.IncludeFile(fileName, false)
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
func listCommand(s *Sqlcmd, args []string, line uint) (err error) {
	cmd := ""
	if args != nil {
		if len(args) > 0 {
			cmd = strings.ToLower(strings.TrimSpace(args[0]))
			if len(args) > 1 || (cmd != "color" && cmd != "") {
				return InvalidCommandError("LIST", line)
			}
		}
	}
	output := s.GetOutput()
	if cmd == "color" {
		sample := "select 'literal' as literal, 100 as number from [sys].[tables]"
		clr := color.TextTypeTSql
		if s.Format.IsXmlMode() {
			sample = `<node att="attValue"/><node>value</node>`
			clr = color.TextTypeXml
		}
		// ignoring errors since it's not critical output
		for _, style := range s.colorizer.Styles() {
			_, _ = output.Write([]byte(style + ": "))
			_ = s.colorizer.Write(output, sample, style, clr)
			_, _ = output.Write([]byte(SqlcmdEol))
		}
		return
	}
	if s.batch == nil || s.batch.String() == "" {
		return
	}

	if err = s.colorizer.Write(output, s.batch.String(), s.vars.ColorScheme(), color.TextTypeTSql); err == nil {
		_, err = output.Write([]byte(SqlcmdEol))
	}

	return
}

func connectCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 {
		return InvalidCommandError("CONNECT", line)
	}

	commandArgs := strings.Fields(args[0])

	// Parse flags
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)
	database := flags.String("D", "", "database name")
	username := flags.String("U", "", "user name")
	password := flags.String("P", "", "password")
	loginTimeout := flags.String("l", "", "login timeout")
	authenticationMethod := flags.String("G", "", "authentication method")

	err := flags.Parse(commandArgs[1:])
	//err := flags.Parse(args[1:])
	if err != nil {
		return InvalidCommandError("CONNECT", line)
	}

	connect := *s.Connect
	connect.UserName, _ = resolveArgumentVariables(s, []rune(*username), false)
	connect.Password, _ = resolveArgumentVariables(s, []rune(*password), false)
	connect.Database, _ = resolveArgumentVariables(s, []rune(*database), false)

	timeout, _ := resolveArgumentVariables(s, []rune(*loginTimeout), false)
	if timeout != "" {
		if timeoutSeconds, err := strconv.ParseInt(timeout, 10, 32); err == nil {
			if timeoutSeconds < 0 {
				return InvalidCommandError("CONNECT", line)
			}
			connect.LoginTimeoutSeconds = int(timeoutSeconds)
		}
	}

	connect.AuthenticationMethod = *authenticationMethod

	// Set server name as the first positional argument
	if len(commandArgs) > 0 {
		connect.ServerName, _ = resolveArgumentVariables(s, []rune(commandArgs[0]), false)
	}

	// If no user name is provided we switch to integrated auth
	_ = s.ConnectDb(&connect, s.lineIo == nil)

	// ConnectDb prints connection errors already, and failure to connect is not fatal even with -b option
	return nil
}

func execCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 {
		return InvalidCommandError("EXEC", line)
	}
	cmdLine := strings.TrimSpace(args[0])
	if cmdLine == "" {
		return InvalidCommandError("EXEC", line)
	}
	if cmdLine, err := resolveArgumentVariables(s, []rune(cmdLine), true); err != nil {
		return err
	} else {
		cmd := sysCommand(cmdLine)
		cmd.Stderr = s.GetError()
		cmd.Stdout = s.GetOutput()
		_ = cmd.Run()
	}
	return nil
}

func editCommand(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return InvalidCommandError("ED", line)
	}
	file, err := os.CreateTemp("", "sq*.sql")
	if err != nil {
		return err
	}
	fileName := file.Name()
	defer os.Remove(fileName)
	text := s.batch.String()
	if s.batch.State() == "-" {
		text = fmt.Sprintf("%s%s", text, SqlcmdEol)
	}
	_, err = file.WriteString(text)
	if err != nil {
		return err
	}
	file.Close()
	cmd := sysCommand(s.vars.TextEditor() + " " + `"` + fileName + `"`)
	cmd.Stderr = s.GetError()
	cmd.Stdout = s.GetOutput()
	err = cmd.Run()
	if err != nil {
		return err
	}
	wasEcho := s.echoFileLines
	s.echoFileLines = true
	s.batch.Reset(nil)
	_ = s.IncludeFile(fileName, false)
	s.echoFileLines = wasEcho
	return nil
}

func onerrorCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) == 0 || args[0] == "" {
		return InvalidCommandError("ON ERROR", line)
	}
	params := strings.TrimSpace(args[0])

	if strings.EqualFold(strings.ToLower(params), "exit") {
		s.Connect.ExitOnError = true
	} else if strings.EqualFold(strings.ToLower(params), "ignore") {
		s.Connect.IgnoreError = true
		s.Connect.ExitOnError = false
	} else {
		return InvalidCommandError("ON ERROR", line)
	}
	return nil
}

func xmlCommand(s *Sqlcmd, args []string, line uint) error {
	if len(args) != 1 || args[0] == "" {
		return InvalidCommandError("XML", line)
	}
	params := strings.TrimSpace(args[0])
	// "OFF" and "ON" are documented as the allowed values.
	// ODBC sqlcmd treats any value other than "ON" the same as "OFF".
	// So we will too.
	if strings.EqualFold(params, "on") {
		s.Format.XmlMode(true)
	} else {
		s.Format.XmlMode(false)
	}
	return nil
}

// helpCommand displays the list of available sqlcmd commands
func helpCommand(s *Sqlcmd, args []string, line uint) error {
	helpText := `:!! [<command>]
  - Executes a command in the operating system shell.
:connect server[\instance] [-l timeout] [-U user [-P password]]
  - Connects to a SQL Server instance.
:ed
  - Edits the current or last executed statement cache.
:error <dest>
  - Redirects error output to a file, stderr, or stdout.
:exit
  - Quits sqlcmd immediately.
:exit()
  - Execute statement cache; quit with no return value.
:exit(<query>)
  - Execute the specified query; returns numeric result.
go [<n>]
  - Executes the statement cache (n times).
:help
  - Shows this list of commands.
:list
  - Prints the content of the statement cache.
:listvar
  - Lists the set sqlcmd scripting variables.
:on error [exit|ignore]
  - Action for batch or sqlcmd command errors.
:out <filename>|stderr|stdout
  - Redirects query output to a file, stderr, or stdout.
:quit
  - Quits sqlcmd immediately.
:r <filename>
  - Append file contents to the statement cache.
:reset
  - Discards the statement cache.
:setvar {variable}
  - Removes a sqlcmd scripting variable.
:setvar <variable> <value>
  - Sets a sqlcmd scripting variable.
:xml [on|off]
  - Sets XML output mode.
`
	_, err := s.GetOutput().Write([]byte(helpText))
	return err
}

func resolveArgumentVariables(s *Sqlcmd, arg []rune, failOnUnresolved bool) (string, error) {
	var b *strings.Builder
	end := len(arg)
	for i := 0; i < end && !s.Connect.DisableVariableSubstitution; {
		c, next := arg[i], grab(arg, i+1, end)
		switch {
		case c == '$' && next == '(':
			vl, ok := readVariableReference(arg, i+2, end)
			if ok {
				varName := string(arg[i+2 : vl])
				val, ok := s.resolveVariable(varName)
				if ok {
					if b == nil {
						b = new(strings.Builder)
						b.Grow(len(arg))
						b.WriteString(string(arg[0:i]))
					}
					b.WriteString(val)
				} else {
					if failOnUnresolved {
						return "", UndefinedVariable(varName)
					}
					s.WriteError(s.GetError(), UndefinedVariable(varName))
					if b != nil {
						b.WriteString(string(arg[i : vl+1]))
					}
				}
				i += ((vl - i) + 1)
			} else {
				if b != nil {
					b.WriteString("$(")
				}
				i += 2
			}
		default:
			if b != nil {
				b.WriteRune(c)
			}
			i++
		}
	}
	if b == nil {
		return string(arg), nil
	}
	return b.String(), nil
}
