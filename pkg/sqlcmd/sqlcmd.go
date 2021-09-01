// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	osuser "os/user"
	"syscall"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/xo/usql/rline"
)

var (
	// ErrExitRequested tells the hosting application to exit immediately
	ErrExitRequested = errors.New("exit")
	// ErrNeedPassword indicates the user should provide a password to enable the connection
	ErrNeedPassword = errors.New("need password")
	// ErrCtrlC indicates execution was ended by ctrl-c or ctrl-break
	ErrCtrlC = errors.New(WarningPrefix + "The last operation was terminated because the user pressed CTRL+C")
)

// ConnectSettings are the settings for connections that can't be
// inferred from scripting variables
type ConnectSettings struct {
	UseTrustedConnection   bool
	TrustServerCertificate bool
}

// Sqlcmd is the core processor for text lines.
//
// It accumulates non-command lines in a buffer and  and sends command lines to the appropriate command runner.
// When the batch delimiter is encountered it sends the current batch to the active connection and prints
// the results to the output writer
type Sqlcmd struct {
	lineIo           rline.IO
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
}

// New creates a new Sqlcmd instance
func New(l rline.IO, workingDirectory string, vars *Variables) *Sqlcmd {
	return &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		batch:            NewBatch(l.Next),
		vars:             vars,
	}
}

// Run processes all available batches.
// When once is true it stops after the first query runs.
func (s *Sqlcmd) Run(once bool) error {
	setupCloseHandler(s)
	stderr, iactive := s.GetError(), s.lineIo.Interactive()
	var lastError error
	for {
		var execute bool
		if iactive {
			s.lineIo.Prompt(s.Prompt())
		}
		var cmd *Command
		var args []string
		var err error
		if s.Query != "" {
			cmd = Commands["GO"]
			args = make([]string, 0)
		} else {
			cmd, args, err = s.batch.Next()
		}
		switch {
		case err == rline.ErrInterrupt:
			// Ignore any error printing the ctrl-c notice since we are exiting
			_, _ = s.GetOutput().Write([]byte(ErrCtrlC.Error() + SqlcmdEol))
			return nil
		case err != nil:
			if err == io.EOF {
				if s.batch.Length == 0 {
					return lastError
				}
				execute = true
			} else {
				_, _ = s.GetOutput().Write([]byte(err.Error() + SqlcmdEol))
			}
		}
		if cmd != nil {
			err = s.RunCommand(cmd, args)
			if err == ErrExitRequested || once {
				s.SetOutput(nil)
				s.SetError(nil)
				break
			}
			if err != nil {
				fmt.Fprintln(stderr, err)
				lastError = err
				continue
			}
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
		return s.lineIo.Stdout()
	}
	return s.out
}

// SetOutput sets the io.WriteCloser to use for non-error output
func (s *Sqlcmd) SetOutput(o io.WriteCloser) {
	if s.out != nil {
		s.out.Close()
	}
	s.out = o
}

// GetError returns the io.Writer to use for errors
func (s *Sqlcmd) GetError() io.Writer {
	if s.err == nil {
		return s.lineIo.Stderr()
	}
	return s.err
}

// SetError sets the io.WriteCloser to use for errors
func (s *Sqlcmd) SetError(e io.WriteCloser) {
	if s.err != nil {
		s.err.Close()
	}
	s.err = e
}

// ConnectionString returns the go-mssql connection string to use for queries
func (s *Sqlcmd) ConnectionString() (connectionString string, err error) {
	serverName, instance, port, err := s.vars.SQLCmdServer()
	if serverName == "" {
		serverName = "."
	}
	if err != nil {
		return "", err
	}
	query := url.Values{}
	connectionURL := &url.URL{
		Scheme: "sqlserver",
		Path:   instance,
	}
	useTrustedConnection := s.Connect.UseTrustedConnection || (s.vars.SQLCmdUser() == "" && !s.vars.UseAad())
	if !useTrustedConnection {
		connectionURL.User = url.UserPassword(s.vars.SQLCmdUser(), s.vars.Password())
	}
	if port > 0 {
		connectionURL.Host = fmt.Sprintf("%s:%d", serverName, port)
	} else {
		connectionURL.Host = serverName
	}
	if s.vars.SQLCmdDatabase() != "" {
		query.Add("database", s.vars.SQLCmdDatabase())
	}

	if s.Connect.TrustServerCertificate {
		query.Add("trustservercertificate", "true")
	}
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String(), nil
}

// ConnectDb opens a connection to the database with the given modifications to the connection
func (s *Sqlcmd) ConnectDb(server string, user string, password string, nopw bool) error {
	if user != "" && password == "" && !nopw {
		return ErrNeedPassword
	}

	connstr, err := s.ConnectionString()
	if err != nil {
		return err
	}

	connectionURL, err := url.Parse(connstr)
	if err != nil {
		return err
	}

	if server != "" {
		serverName, instance, port, err := splitServer(server)
		if err != nil {
			return err
		}
		connectionURL.Path = instance
		if port > 0 {
			connectionURL.Host = fmt.Sprintf("%s:%d", serverName, port)
		} else {
			connectionURL.Host = serverName
		}
	}

	if password == "" {
		password = s.vars.Password()
	}

	if user != "" {
		connectionURL.User = url.UserPassword(user, password)
	}

	connector, err := mssql.NewConnector(connectionURL.String())
	if err != nil {
		return err
	}
	db := sql.OpenDB(connector)
	err = db.Ping()
	if err != nil {
		return err
	}
	// we got a good connection so we can update the Sqlcmd
	if s.db != nil {
		s.db.Close()
	}
	s.db = db
	if server != "" {
		s.vars.Set(SQLCMDSERVER, server)
	}
	if user != "" {
		s.vars.Set(SQLCMDUSER, user)
		s.Connect.UseTrustedConnection = false
		if password != "" {
			s.vars.Set(SQLCMDPASSWORD, password)
		}
	} else if s.vars.SQLCmdUser() == "" {
		u, e := osuser.Current()
		if e != nil {
			panic("Unable to get user name")
		}
		s.Connect.UseTrustedConnection = true
		s.vars.Set(SQLCMDUSER, u.Username)
	}

	if s.batch != nil {
		s.batch.batchline = 1
	}
	return nil
}

func setupCloseHandler(s *Sqlcmd) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_, _ = s.GetOutput().Write([]byte(ErrCtrlC.Error()))
		os.Exit(0)
	}()
}
