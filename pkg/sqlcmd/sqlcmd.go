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
	"time"

	"github.com/golang-sql/sqlexp"
	mssql "github.com/microsoft/go-mssqldb"

	_ "github.com/microsoft/go-mssqldb/aecmk/akv"
	_ "github.com/microsoft/go-mssqldb/aecmk/localcert"
	"github.com/microsoft/go-mssqldb/msdsn"
	"github.com/microsoft/go-sqlcmd/internal/color"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	// ErrExitRequested tells the hosting application to exit immediately
	ErrExitRequested = errors.New("exit")
	// ErrNeedPassword indicates the user should provide a password to enable the connection
	ErrNeedPassword = errors.New("need password")
	// ErrCtrlC indicates execution was ended by ctrl-c or ctrl-break
	ErrCtrlC = errors.New(WarningPrefix + "The last operation was terminated because the user pressed CTRL+C")
	// ErrCommandsDisabled indicates system commands and startup script are disabled
	ErrCommandsDisabled = &CommonSqlcmdErr{
		message: ErrCmdDisabled,
	}
)

// Console defines methods used for console input and output
type Console interface {
	// Readline returns the next line of input.
	Readline() (string, error)
	// Readpassword displays the given prompt and returns a password
	ReadPassword(prompt string) ([]byte, error)
	// SetPrompt sets the prompt text shown to input the next line
	SetPrompt(s string)
	// Close clears any buffers and closes open file handles
	Close()
}

// Sqlcmd is the core processor for text lines.
//
// It accumulates non-command lines in a buffer and sends command lines to the appropriate command runner.
// When the batch delimiter is encountered it sends the current batch to the active connection and prints
// the results to the output writer
type Sqlcmd struct {
	lineIo           Console
	workingDirectory string
	db               *sql.Conn
	out              io.WriteCloser
	err              io.WriteCloser
	batch            *Batch
	echoFileLines    bool
	// Exitcode is returned to the operating system when the process exits
	Exitcode int
	// Connect controls how Sqlcmd connects to the database
	Connect *ConnectSettings
	vars    *Variables
	// Format renders the query output
	Format Formatter
	// Query is the TSQL query to run
	Query string
	// Cmd provides the implementation of commands like :list and GO
	Cmd Commands
	// PrintError allows the host to redirect errors away from the default output. Returns false if the error is not redirected by the host.
	PrintError func(msg string, severity uint8) bool
	// UnicodeOutputFile is true when UTF16 file output is needed
	UnicodeOutputFile bool
	// EchoInput tells the GO command to print the batch text before running the query
	EchoInput bool
	colorizer color.Colorizer
	termchan  chan os.Signal
}

// New creates a new Sqlcmd instance.
// The Console instane must be non-nil for Sqlcmd to run in interactive mode.
// The hosting application is responsible for calling Close() on the Console instance before process exit.
func New(l Console, workingDirectory string, vars *Variables) *Sqlcmd {
	s := &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		vars:             vars,
		Cmd:              newCommands(),
		Connect:          &ConnectSettings{},
		colorizer:        color.New(false),
	}
	s.batch = NewBatch(s.scanNext, s.Cmd)
	s.batch.ParseVariables = func() bool { return !s.Connect.DisableVariableSubstitution }
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
// The error returned from Run is mainly of informational value. Its Message will have been printed
// before Run returns.
func (s *Sqlcmd) Run(once bool, processAll bool) error {
	iactive := s.lineIo != nil
	var lastError error
	for {
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

				if !processAll {
					return nil
				}
				// Run the GO and exit
				cmd = s.Cmd["GO"]
				args = make([]string, 0)
				once = true
			} else {
				s.WriteError(s.GetOutput(), err)
			}
		}
		if cmd != nil {
			lastError = nil
			err = s.RunCommand(cmd, args)
			if err == ErrExitRequested || once {
				break
			}
			if err != nil {
				s.WriteError(s.GetOutput(), err)
				lastError = err
			}
		}

		// Some Console implementations catch the ctrl-c so s.termchan isn't signalled
		if err == ErrCtrlC {
			s.Exitcode = 0
			return err
		}
		if err != nil && err != io.EOF && (s.Connect.ExitOnError && !s.Connect.IgnoreError) {
			// If the error were due to a SQL error, the GO command handler
			// would have set ExitCode already
			if s.Exitcode == 0 {
				s.Exitcode = 1
			}
			lastError = err
			break
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

// WriteError writes the error on specified stream
func (s *Sqlcmd) WriteError(stream io.Writer, err error) {
	if serr, ok := err.(SqlcmdError); ok {
		if s.GetError() != os.Stdout {
			_, _ = s.GetError().Write([]byte(serr.Error() + SqlcmdEol))
		} else {
			_, _ = os.Stderr.Write([]byte(serr.Error() + SqlcmdEol))
		}
	} else {
		_, _ = stream.Write([]byte(err.Error() + SqlcmdEol))
	}
}

// ConnectDb opens a connection to the database with the given modifications to the connection
// nopw == true means don't prompt for a password if the auth type requires it
// if connect is nil, ConnectDb uses the current connection. If non-nil and the connection succeeds,
// s.Connect is replaced with the new value.
func (s *Sqlcmd) ConnectDb(connect *ConnectSettings, nopw bool) error {
	newConnection := connect != nil
	if connect == nil {
		connect = s.Connect
	}

	var connector driver.Connector
	useAad := !connect.sqlAuthentication() && !connect.integratedAuthentication()
	if connect.RequiresPassword() && !nopw && connect.Password == "" {
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
	db, err := sql.OpenDB(connector).Conn(context.Background())
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
		// osuser.Current() returns an error in some restricted environments
		if e == nil {
			s.vars.Set(SQLCMDUSER, u.Username)
		}
	}
	if newConnection {
		s.Connect = connect
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
	pwd, err := s.lineIo.ReadPassword(localizer.Sprintf("Password:"))
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
		return InvalidFileError(err, path)
	}
	defer f.Close()
	b := s.batch.batchline
	utf16bom := unicode.BOMOverride(unicode.UTF8.NewDecoder())
	unicodeReader := transform.NewReader(f, utf16bom)
	scanner := bufio.NewReader(unicodeReader)
	curLine := s.batch.read
	echoFileLines := s.echoFileLines
	ln := make([]byte, 0, 2*1024*1024)
	s.batch.read = func() (string, error) {
		var (
			isPrefix bool  = true
			err      error = nil
			line     []byte
		)

		for isPrefix && err == nil {
			line, isPrefix, err = scanner.ReadLine()
			ln = append(ln, line...)
		}
		if err == nil && echoFileLines {
			_, _ = s.GetOutput().Write([]byte(s.Prompt()))
			_, _ = s.GetOutput().Write(ln)
			_, _ = s.GetOutput().Write([]byte(SqlcmdEol))
		}
		t := string(ln)
		ln = ln[:0]
		return t, err
	}
	err = s.Run(false, processAll)
	s.batch.read = curLine
	if !s.echoFileLines {
		if s.batch.State() == "=" {
			s.batch.batchline = 1
		} else {
			s.batch.batchline = b + 1
		}
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
	// The varmap index is rune based not byte based
	r := []rune(q)
	keys := make([]int, 0, len(s.batch.varmap))
	for k := range s.batch.varmap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	last := 0
	for _, i := range keys {
		b.WriteString(string(r[last:i]))
		v := s.batch.varmap[i]
		if val, ok := s.resolveVariable(v); ok {
			b.WriteString(val)
		} else {
			_, _ = fmt.Fprintf(s.GetError(), "'%s' scripting variable not defined.%s", v, SqlcmdEol)
			b.WriteString(fmt.Sprintf("$(%s)", v))
		}
		last = i + len([]rune(v)) + 3
	}
	b.WriteString(string(r[last:]))
	return b.String()
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
	timeout := s.vars.QueryTimeoutSeconds()
	if timeout > 0 {
		ct, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		ctx = ct
	}
	retmsg := &sqlexp.ReturnMessage{}
	rows, qe := s.db.QueryContext(ctx, query, retmsg)
	if rows != nil {
		defer func() { _ = rows.Close() }()
	}
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
			if !s.PrintError(m.Message.String(), 10) {
				s.Format.AddMessage(m.Message.String())
				switch e := m.Message.(type) {
				case mssql.Error:
					qe = s.handleError(&retcode, e)
				}
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
				s.Format.AddMessage(localizer.Sprintf("(1 row affected)"))
			} else {
				s.Format.AddMessage(localizer.Sprintf("(%d rows affected)", m.Count))
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
			inresult := rows.Next()
			for inresult {
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
	var errState uint8
	var errNumber int32
	switch sqlError := err.(type) {
	case mssql.Error:
		errSeverity = sqlError.Class
		errState = sqlError.State
		errNumber = sqlError.Number
	}

	// 127 is the magic exit code
	if errState == 127 {
		*retcode = int(errNumber)
		return ErrExitRequested
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

// SetupCloseHandler subscribes to the os.Signal channel for SIGTERM.
// When it receives the event due to the user pressing ctrl-c or ctrl-break
// that isn't handled directly by the Console or hosting application,
// it will call Close() on the Console and exit the application.
// Use StopCloseHandler to remove the subscription
func (s *Sqlcmd) SetupCloseHandler() {
	s.termchan = make(chan os.Signal, 1)
	signal.Notify(s.termchan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-s.termchan
		s.WriteError(s.GetOutput(), ErrCtrlC)
		if s.lineIo != nil {
			s.lineIo.Close()
		}
		os.Exit(0)
	}()
}

// StopCloseHandler unsubscribes the Sqlcmd from the SIGTERM signal
func (s *Sqlcmd) StopCloseHandler() {
	signal.Stop(s.termchan)
}
