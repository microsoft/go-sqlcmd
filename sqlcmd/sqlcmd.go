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
	"github.com/microsoft/go-sqlcmd/sqlcmderrors"
	"github.com/microsoft/go-sqlcmd/util"
	"github.com/microsoft/go-sqlcmd/variables"
	"github.com/xo/usql/rline"
)

var (
	ErrExitRequested = errors.New("exit")
	ErrNeedPassword  = errors.New("need password")
	ErrCtrlC         = errors.New(sqlcmderrors.WarningPrefix + "The last operation was terminated because the user pressed CTRL+C")
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
	vars     *variables.Variables
	Format   Formatter
	Query    string
}

// New creates a new Sqlcmd instance
func New(l rline.IO, workingDirectory string, vars *variables.Variables) *Sqlcmd {
	return &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		batch:            NewBatch(l.Next),
		vars:             vars,
	}
}

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
			s.GetOutput().Write([]byte(ErrCtrlC.Error()))
			return nil
		case err != nil:
			if err == io.EOF {
				if s.batch.Length == 0 {
					return lastError
				} else {
					execute = true
				}
			} else {
				return err
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
			fmt.Fprintln(s.GetOutput(), "Execute: "+s.batch.String())
			s.batch.Reset(nil)
		}
	}
	return lastError
}

func (s *Sqlcmd) Prompt() string {
	ch := ">"
	if s.batch.quote != 0 || s.batch.comment {
		ch = "~"
	}
	return fmt.Sprint(s.batch.batchline) + ch + " "
}

func (s *Sqlcmd) RunCommand(cmd *Command, args []string) error {
	return cmd.action(s, args, s.batch.linecount)
}

func (s *Sqlcmd) GetOutput() io.Writer {
	if s.out == nil {
		return s.lineIo.Stdout()
	}
	return s.out
}

func (s *Sqlcmd) SetOutput(o io.WriteCloser) {
	if s.out != nil {
		s.out.Close()
	}
	s.out = o
}

func (s *Sqlcmd) GetError() io.Writer {
	if s.err == nil {
		return s.lineIo.Stderr()
	}
	return s.err
}

func (s *Sqlcmd) SetError(e io.WriteCloser) {
	if s.err != nil {
		s.err.Close()
	}
	s.err = e
}

func (s *Sqlcmd) ConnectionString() (connectionString string, err error) {
	serverName, instance, port, err := s.vars.SqlCmdServer()
	if serverName == "" {
		serverName = "."
	}
	if err != nil {
		return "", err
	}
	query := url.Values{}
	connectionUrl := &url.URL{
		Scheme: "sqlserver",
		Path:   instance,
	}
	useTrustedConnection := s.Connect.UseTrustedConnection || (s.vars.SqlCmdUser() == "" && !s.vars.UseAad())
	if !useTrustedConnection {
		connectionUrl.User = url.UserPassword(s.vars.SqlCmdUser(), s.vars.Password())
	}
	if port > 0 {
		connectionUrl.Host = fmt.Sprintf("%s:%d", serverName, port)
	} else {
		connectionUrl.Host = serverName
	}
	if s.vars.SqlCmdDatabase() != "" {
		query.Add("database", s.vars.SqlCmdDatabase())
	}

	if s.Connect.TrustServerCertificate {
		query.Add("trustservercertificate", "true")
	}
	connectionUrl.RawQuery = query.Encode()
	return connectionUrl.String(), nil
}

// Opens a connection to the database with the given modifications to the connection
func (s *Sqlcmd) ConnectDb(server string, user string, password string, nopw bool) error {
	if user != "" && password == "" && !nopw {
		return ErrNeedPassword
	}

	connstr, err := s.ConnectionString()
	if err != nil {
		return err
	}

	connectionUrl, err := url.Parse(connstr)
	if err != nil {
		return err
	}

	if server != "" {
		serverName, instance, port, err := util.SplitServer(server)
		if err != nil {
			return err
		}
		connectionUrl.Path = instance
		if port > 0 {
			connectionUrl.Host = fmt.Sprintf("%s:%d", serverName, port)
		} else {
			connectionUrl.Host = serverName
		}
	}

	if password == "" {
		password = s.vars.Password()
	}

	if user != "" {
		connectionUrl.User = url.UserPassword(user, password)
	}

	connector, err := mssql.NewConnector(connectionUrl.String())
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
		s.vars.Set(variables.SQLCMDSERVER, server)
	}
	if user != "" {
		s.vars.Set(variables.SQLCMDUSER, user)
		s.Connect.UseTrustedConnection = false
		if password != "" {
			s.vars.Set(variables.SQLCMDPASSWORD, password)
		}
	} else if s.vars.SqlCmdUser() == "" {
		u, e := osuser.Current()
		if e != nil {
			panic("Unable to get user name")
		}
		s.Connect.UseTrustedConnection = true
		s.vars.Set(variables.SQLCMDUSER, u.Username)
	}

	if s.batch != nil {
		s.batch.batchline = 1
	}
	return nil
}

func setupCloseHandler(s *Sqlcmd) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		s.GetOutput().Write([]byte(ErrCtrlC.Error()))
		os.Exit(0)
	}()
}
