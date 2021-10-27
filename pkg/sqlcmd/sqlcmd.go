// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bufio"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	osuser "os/user"
	"sort"
	"strings"
	"syscall"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/gohxs/readline"
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
	// UseTrustedConnection indicates integrated auth is used when no user name is provided
	UseTrustedConnection bool
	// TrustServerCertificate sets the TrustServerCertificate setting on the connection string
	TrustServerCertificate bool
	AuthenticationMethod   string
	// DisableEnvironmentVariables determines if sqlcmd resolves scripting variables from the process environment
	DisableEnvironmentVariables bool
	// DisableVariableSubstitution determines if scripting variables should be evaluated
	DisableVariableSubstitution bool
	// Password is the password used with SQL authentication
	Password string
}

func (c ConnectSettings) authenticationMethod() string {
	if c.AuthenticationMethod == "" {
		return NotSpecified
	}
	return c.AuthenticationMethod
}

// Sqlcmd is the core processor for text lines.
//
// It accumulates non-command lines in a buffer and  and sends command lines to the appropriate command runner.
// When the batch delimiter is encountered it sends the current batch to the active connection and prints
// the results to the output writer
type Sqlcmd struct {
	lineIo           *readline.Instance
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
}

func EnterNewPassword(l *readline.Instance) (string, error) {
	pwchars, pwerr := l.ReadPassword("Password: ")
	return string(pwchars), pwerr
}

// New creates a new Sqlcmd instance
func New(l *readline.Instance, workingDirectory string, vars *Variables) *Sqlcmd {
	s := &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		vars:             vars,
		Cmd:              newCommands(),
	}
	s.batch = NewBatch(s.scanNext, s.Cmd)
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
	stderr, iactive := s.GetError(), s.lineIo != nil
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
		switch {
		case err == readline.ErrInterrupt:
			// Ignore any error printing the ctrl-c notice since we are exiting
			_, _ = s.GetOutput().Write([]byte(ErrCtrlC.Error() + SqlcmdEol))
			return nil
		case err != nil:
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
		return os.Stdout
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
		return os.Stderr
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

	if s.sqlAuthentication() {
		connectionURL.User = url.UserPassword(s.vars.SQLCmdUser(), s.Connect.Password)
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

	// read environment variables for any parameters that were not supplied
	if user == "" {
		user = s.vars.SQLCmdUser()
	}
	if password == "" {
		password = s.Connect.Password
	}

	if user != "" && password == "" && !nopw {
		// user was specified and pw is required so propt user for it
		// if SA is desired but no password specified, query for it
		newpassword, prompterr := EnterNewPassword(s.lineIo)
		if prompterr != nil {
			return ErrNeedPassword
		}

		password = newpassword
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

	var connector driver.Connector
	// To determine whether to use Sql auth/windows auth/aad auth, compare the current ConnectSettings with the new parameters
	// If sqlcmd was started with sql auth or windows auth, :connect will not switch to AAD
	// if sqlcmd was started with AAD auth, it will remain in some variant of AAD auth depending on the user/password combination
	useAad := !s.sqlAuthentication() && !s.integratedAuthentication()
	if password == "" {
		password = s.Connect.Password
	}
	if !useAad {
		if user != "" {
			connectionURL.User = url.UserPassword(user, password)
		}

		connector, err = mssql.NewConnector(connectionURL.String())
	} else {
		if user == "" {
			user = s.vars.SQLCmdUser()
		}
		connector, err = s.GetTokenBasedConnection(connectionURL.String(), user, password)
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
	if server != "" {
		s.vars.Set(SQLCMDSERVER, server)
	}
	if user != "" {
		s.vars.Set(SQLCMDUSER, user)
		s.Connect.UseTrustedConnection = false
		s.Connect.Password = password
	} else if s.vars.SQLCmdUser() == "" {
		u, e := osuser.Current()
		if e != nil {
			panic("Unable to get user name")
		}
		if !useAad {
			s.Connect.UseTrustedConnection = true
		}
		s.vars.Set(SQLCMDUSER, u.Username)
	}

	if s.batch != nil {
		s.batch.batchline = 1
	}
	return nil
}

// IncludeFile opens the given file and processes its batches
// When processAll is true text not followed by a go statement is run as a query
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

func (s *Sqlcmd) integratedAuthentication() bool {
	return s.Connect.UseTrustedConnection || (s.vars.SQLCmdUser() == "" && s.Connect.authenticationMethod() == NotSpecified)
}

func (s *Sqlcmd) sqlAuthentication() bool {
	return s.Connect.authenticationMethod() == SqlPassword ||
		(!s.Connect.UseTrustedConnection && s.Connect.authenticationMethod() == NotSpecified && s.vars.SQLCmdUser() != "")
}

// runQuery runs the query and prints the results
// The return value is based on the first cell of the last column of the last result set.
// If it's numeric, it will be converted to int
// -100 : Error encountered prior to selecting return value
// -101: No rows found
// -102: Conversion error occurred when selecting return value
func (s *Sqlcmd) runQuery(query string) int {
	retcode := -101
	s.Format.BeginBatch(query, s.vars, s.GetOutput(), s.GetError())
	rows, qe := s.db.Query(query)
	if qe != nil {
		s.Format.AddError(qe)
	}
	var err error
	var cols []*sql.ColumnType
	results := true
	for qe == nil && results {
		cols, err = rows.ColumnTypes()
		if err != nil {
			retcode = -100
			s.Format.AddError(err)
		} else {
			s.Format.BeginResultSet(cols)
			active := rows.Next()
			for active {
				col1 := s.Format.AddRow(rows)
				active = rows.Next()
				if !active {
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
					s.Format.AddError(err)
				}
			}
			s.Format.EndResultSet()
		}
		results = rows.NextResultSet()
		if err = rows.Err(); err != nil {
			retcode = -100
			s.Format.AddError(err)
		}
	}
	s.Format.EndBatch()
	return retcode
}
