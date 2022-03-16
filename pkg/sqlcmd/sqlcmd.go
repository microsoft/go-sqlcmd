// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bufio"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	osuser "os/user"
	"sort"
	"strings"
	"syscall"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/denisenkom/go-mssqldb/msdsn"
	"github.com/golang-sql/sqlexp"
)

var (
	// ErrExitRequested tells the hosting application to exit immediately
	ErrExitRequested = errors.New("exit")
	// ErrNeedPassword indicates the user should provide a password to enable the connection
	ErrNeedPassword = errors.New("need password")
	// ErrCtrlC indicates execution was ended by ctrl-c or ctrl-break
	ErrCtrlC = errors.New(WarningPrefix + "The last operation was terminated because the user pressed CTRL+C")
)

// Console defines methods used for console input and output
type Console interface {
	// Readline returns the next line of input.
	Readline() (string, error)
	// Readpassword displays the given prompt and returns a password
	ReadPassword(prompt string) ([]byte, error)
	// SetPrompt sets the prompt text shown to input the next line
	SetPrompt(s string)
}

// Sqlcmd is the core processor for text lines.
//
// It accumulates non-command lines in a buffer and  and sends command lines to the appropriate command runner.
// When the batch delimiter is encountered it sends the current batch to the active connection and prints
// the results to the output writer
type Sqlcmd struct {
	lineIo           Console
	workingDirectory string
	db               *sql.DB
	out              io.WriteCloser
	err              io.WriteCloser
	batch            *Batch
	// Exitcode is returned to the operating system when the process exits
	Exitcode int
	Connect  ConnectSettings
	vars     *Variables
	Format   Formatter
	Query    string
	Cmd      Commands
	// PrintError allows the host to redirect errors away from the default output. Returns false if the error is not redirected by the host.
	PrintError func(msg string, severity uint8) bool
}

// New creates a new Sqlcmd instance
func New(l Console, workingDirectory string, vars *Variables) *Sqlcmd {
	s := &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		vars:             vars,
		Cmd:              newCommands(),
	}
	s.batch = NewBatch(s.scanNext, s.Cmd)
	mssql.SetContextLogger(s)
	s.PrintError = func(msg string, severity uint8) bool {
		return false
	}
	return s
}

func (s *Sqlcmd) scanNext() (string, error) {
	return s.lineIo.Readline()
}

// Run processes all available batches.
// When once is true it stops after the first query runs.
// When processAll is true it executes any remaining batch content when reaching EOF
func (s *Sqlcmd) Run(once bool, processAll bool) error {
	setupCloseHandler(s)
	iactive := s.lineIo != nil
	var lastError error
	for {
		var execute bool
		if iactive {
			s.lineIo.SetPrompt(s.Prompt())
		}
		var cmd *Command
		var args []string
		var err error
		if s.Query != "" {
			s.batch.Reset([]rune(s.Query))
			// batch.Next validates variable syntax
			cmd, args, err = s.batch.Next()
			if cmd == nil {
				cmd = s.Cmd["GO"]
				args = make([]string, 0)
			}
			s.Query = ""
		} else {
			cmd, args, err = s.batch.Next()
		}

		if err != nil {
			if err == io.EOF {
				if s.batch.Length == 0 {
					return lastError
				}
				execute = processAll
				if !execute {
					return nil
				}
			} else {
				_, _ = s.GetOutput().Write([]byte(err.Error() + SqlcmdEol))
			}
		}
		if cmd != nil {
			lastError = nil
			err = s.RunCommand(cmd, args)
			if err == ErrExitRequested || once {
				break
			}
			if err != nil {
				_, _ = s.GetOutput().Write([]byte(err.Error() + SqlcmdEol))
				lastError = err
			}
		}
		if err != nil && s.Connect.ExitOnError {
			// If the error were due to a SQL error, the GO command handler
			// would have set ExitCode already
			if s.Exitcode == 0 {
				s.Exitcode = 1
			}
			lastError = err
			break
		}
		if execute {
			s.Query = s.batch.String()
			once = true
			s.batch.Reset(nil)
		}
	}
	return lastError
}

// Prompt returns the current user prompt message
func (s *Sqlcmd) Prompt() string {
	ch := ">"
	if s.batch.quote != 0 || s.batch.comment {
		ch = "~"
	}
	return fmt.Sprint(s.batch.batchline) + ch + " "
}

// RunCommand performs the given Command
func (s *Sqlcmd) RunCommand(cmd *Command, args []string) error {
	return cmd.action(s, args, s.batch.linecount)
}

// GetOutput returns the io.Writer to use for non-error output
func (s *Sqlcmd) GetOutput() io.Writer {
	if s.out == nil {
		return os.Stdout
	}
	return s.out
}

// SetOutput sets the io.WriteCloser to use for non-error output
func (s *Sqlcmd) SetOutput(o io.WriteCloser) {
	if s.out != nil && s.out != os.Stderr && s.out != os.Stdout {
		s.out.Close()
	}
	s.out = o
}

// GetError returns the io.Writer to use for errors
func (s *Sqlcmd) GetError() io.Writer {
	if s.err == nil {
		return s.GetOutput()
	}
	return s.err
}

// SetError sets the io.WriteCloser to use for errors
func (s *Sqlcmd) SetError(e io.WriteCloser) {
	if s.err != nil && s.err != os.Stderr && s.err != os.Stdout {
		s.err.Close()
	}
	s.err = e
}

// ConnectDb opens a connection to the database with the given modifications to the connection
// nopw == true means don't prompt for a password if the auth type requires it
// if connect is nil, ConnectDb uses the current connection. If non-nil and the connection succeeds,
// s.Connect is replaced with the new value.
func (s *Sqlcmd) ConnectDb(connect *ConnectSettings, nopw bool) error {
	newConnection := connect != nil
	if connect == nil {
		connect = &s.Connect
	}

	var connector driver.Connector
	useAad := !connect.sqlAuthentication() && !connect.integratedAuthentication()
	if connect.requiresPassword() && !nopw && connect.Password == "" {
		var err error
		if connect.Password, err = s.promptPassword(); err != nil {
			return err
		}
	}
	connstr, err := connect.ConnectionString()
	if err != nil {
		return err
	}

	if !useAad {
		connector, err = mssql.NewConnector(connstr)
	} else {
		connector, err = GetTokenBasedConnection(connstr, connect.authenticationMethod())
	}
	if err != nil {
		return err
	}
	db := sql.OpenDB(connector)
	err = db.Ping()
	if err != nil {
		fmt.Fprintln(s.GetOutput(), err)
		return err
	}
	// we got a good connection so we can update the Sqlcmd
	if s.db != nil {
		s.db.Close()
	}
	s.db = db
	s.vars.Set(SQLCMDSERVER, connect.ServerName)
	s.vars.Set(SQLCMDDBNAME, connect.Database)
	if connect.UserName != "" {
		s.vars.Set(SQLCMDUSER, connect.UserName)
	} else {
		u, e := osuser.Current()
		if e != nil {
			panic("Unable to get user name")
		}
		s.vars.Set(SQLCMDUSER, u.Username)
	}
	if newConnection {
		s.Connect = *connect
	}
	if s.batch != nil {
		s.batch.batchline = 1
	}
	return nil
}

func (s *Sqlcmd) promptPassword() (string, error) {
	if s.lineIo == nil {
		return "", nil
	}
	pwd, err := s.lineIo.ReadPassword("Password:")
	if err != nil {
		return "", err
	}

	return string(pwd), nil
}

// IncludeFile opens the given file and processes its batches.
// When processAll is true, text not followed by a go statement is run as a query
func (s *Sqlcmd) IncludeFile(path string, processAll bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := s.batch.batchline
	scanner := bufio.NewScanner(f)
	curLine := s.batch.read
	s.batch.read = func() (string, error) {
		if !scanner.Scan() {
			err := scanner.Err()
			if err == nil {
				return "", io.EOF
			}
			return "", err
		}
		return scanner.Text(), nil
	}
	err = s.Run(false, processAll)
	s.batch.read = curLine
	if s.batch.State() == "=" {
		s.batch.batchline = 1
	} else {
		s.batch.batchline = b + 1
	}
	return err
}

// resolveVariable returns the value of the named variable
func (s *Sqlcmd) resolveVariable(v string) (string, bool) {
	if val, ok := s.vars.Get(v); ok {
		return val, ok
	}

	if !s.Connect.DisableEnvironmentVariables {
		return os.LookupEnv(v)
	}
	return "", false
}

// getRunnableQuery converts the raw batch into an executable query by
// replacing variable references with their resolved values
// If variables are not used, returns the original string
func (s *Sqlcmd) getRunnableQuery(q string) string {
	if s.Connect.DisableVariableSubstitution || len(s.batch.varmap) == 0 {
		return q
	}
	b := new(strings.Builder)
	b.Grow(len(q))
	keys := make([]int, 0, len(s.batch.varmap))
	for k := range s.batch.varmap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	last := 0
	for _, i := range keys {
		b.WriteString(q[last:i])
		v := s.batch.varmap[i]
		if val, ok := s.resolveVariable(v); ok {
			b.WriteString(val)
		} else {
			_, _ = fmt.Fprintf(s.GetError(), "'%s' scripting variable not defined.%s", v, SqlcmdEol)
			b.WriteString(fmt.Sprintf("$(%s)", v))
		}
		last = i + len(v) + 3
	}
	b.WriteString(q[last:])
	return b.String()
}

func setupCloseHandler(s *Sqlcmd) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_, _ = s.GetOutput().Write([]byte(ErrCtrlC.Error() + SqlcmdEol))
		os.Exit(0)
	}()
}

// runQuery runs the query and prints the results
// The return value is based on the first cell of the last column of the last result set.
// If it's numeric, it will be converted to int
// -100 : Error encountered prior to selecting return value
// -101: No rows found
// -102: Conversion error occurred when selecting return value
func (s *Sqlcmd) runQuery(query string) (int, error) {
	retcode := -101
	s.Format.BeginBatch(query, s.vars, s.GetOutput(), s.GetError())
	ctx := context.Background()
	retmsg := &sqlexp.ReturnMessage{}
	rows, qe := s.db.QueryContext(ctx, query, retmsg)
	if qe != nil {
		s.Format.AddError(qe)
	}
	var err error
	var cols []*sql.ColumnType
	results := true
	first := true
	for qe == nil && results {
		msg := retmsg.Message(ctx)
		switch m := msg.(type) {
		case sqlexp.MsgNotice:
			if !s.PrintError(m.Message, 10) {
				s.Format.AddMessage(m.Message)
			}
		case sqlexp.MsgError:
			switch e := m.Error.(type) {
			case mssql.Error:
				if !s.PrintError(e.Message, e.Class) {
					s.Format.AddError(m.Error)
				}
			}
			qe = s.handleError(&retcode, m.Error)
		case sqlexp.MsgRowsAffected:
			if m.Count == 1 {
				s.Format.AddMessage("(1 row affected)")
			} else {
				s.Format.AddMessage(fmt.Sprintf("(%d rows affected)", m.Count))
			}
		case sqlexp.MsgNextResultSet:
			results = rows.NextResultSet()
			if err = rows.Err(); err != nil {
				retcode = -100
				qe = s.handleError(&retcode, err)
				s.Format.AddError(err)
			}
			if results {
				first = true
			}
		case sqlexp.MsgNext:
			inresult := rows.Next()
			for inresult {
				if first {
					first = false
					cols, err = rows.ColumnTypes()
					if err != nil {
						retcode = -100
						qe = s.handleError(&retcode, err)
						s.Format.AddError(err)
					} else {
						s.Format.BeginResultSet(cols)
					}
				}
				col1 := s.Format.AddRow(rows)
				inresult = rows.Next()
				if !inresult {
					if col1 == "" {
						retcode = 0
					} else if _, cerr := fmt.Sscanf(col1, "%d", &retcode); cerr != nil {
						retcode = -102
					}
				}
			}
			if retcode != -102 {
				if err = rows.Err(); err != nil {
					retcode = -100
					qe = s.handleError(&retcode, err)
					s.Format.AddError(err)
				}
			}
			s.Format.EndResultSet()
		}
	}
	s.Format.EndBatch()
	return retcode, qe
}

// returns ErrExitRequested if the error is a SQL error and satisfies the connection's error handling configuration
func (s *Sqlcmd) handleError(retcode *int, err error) error {
	if err == nil {
		return nil
	}

	var minSeverityToExit uint8 = 11
	if s.Connect.ErrorSeverityLevel > 0 {
		minSeverityToExit = s.Connect.ErrorSeverityLevel
	}
	var errSeverity uint8
	switch sqlError := err.(type) {
	case mssql.Error:
		errSeverity = sqlError.Class
	}

	if s.Connect.ErrorSeverityLevel > 0 {
		if errSeverity >= minSeverityToExit {
			*retcode = int(errSeverity)
			s.Exitcode = *retcode
		}
	} else if s.Connect.ExitOnError {
		if errSeverity >= minSeverityToExit {
			*retcode = 1
		}
	}
	if s.Connect.ExitOnError && errSeverity >= minSeverityToExit {
		return ErrExitRequested
	}
	return nil
}

// Log attempts to write driver traces to the current output. It ignores errors
func (s Sqlcmd) Log(_ context.Context, _ msdsn.Log, msg string) {
	_, _ = s.GetOutput().Write([]byte("DRIVER:" + msg))
	_, _ = s.GetOutput().Write([]byte(SqlcmdEol))
}
