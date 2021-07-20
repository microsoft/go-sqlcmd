package sqlcmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/microsoft/go-sqlcmd/variables"
	"github.com/xo/usql/rline"
)

var (
	ErrExitRequested = errors.New("exit")
)

// ConnectSettings are the global settings for all connections that can't be
// overridden by :connect or by scripting variables
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
	//	db               *sql.DB
	out   io.WriteCloser
	err   io.WriteCloser
	batch *Batch
	// Exitcode is returned to the operating system when the process exits
	Exitcode int
	Connect  ConnectSettings
	vars     *variables.Variables
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

func (s *Sqlcmd) Run() error {
	stderr, iactive := s.GetError(), s.lineIo.Interactive()
	var lastError error
	for {
		var execute bool
		if iactive {
			s.lineIo.Prompt(s.Prompt())
		}
		cmd, args, err := s.batch.Next()
		switch {
		case err == rline.ErrInterrupt:
			s.batch.Reset(nil)
			continue
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
			if err == ErrExitRequested {
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
