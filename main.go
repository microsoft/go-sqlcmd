// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	//"database/sql"

	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/microsoft/go-sqlcmd/sqlcmd"
	"github.com/microsoft/go-sqlcmd/variables"
	"github.com/xo/usql/rline"
)

// The exhaustive list is at https://docs.microsoft.com/sql/tools/sqlcmd-utility?view=sql-server-ver15
type SqlCmdArguments struct {
	// Which batch terminator to use. Default is GO
	BatchTerminator string `short:"c" default:"GO" arghelp:"Specifies the batch terminator. The default value is GO."`
	// Whether to trust the server certificate on an encrypted connection
	TrustServerCertificate bool   `short:"C" help:"Implicitly trust the server certificate without validation."`
	DatabaseName           string `short:"d" help:"This option sets the sqlcmd scripting variable SQLCMDDBNAME. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits."`
	UseTrustedConnection   bool   `short:"E" xor:"uid" help:"Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any any environment variables that define user name and password."`
	Password               string `short:"P" xor:"pwd" help:"User-specified password."`
	UserName               string `short:"U" xor:"uid" help:"The login name or contained database user name.  For contained database users, you must provide the database name option"`
	// Files from which to read query text
	InputFile  []string `short:"i" xor:"input1, input2" type:"existingFile" help:"Identifies one or more files that contain batches of SQL statements. If one or more files do not exist, sqlcmd will exit. Mutually exclusive with -Q/-q."`
	OutputFile string   `short:"o" type:"path" help:"Identifies the file that receives output from sqlcmd."`
	// First query to run in interactive mode
	InitialQuery string `short:"q" xor:"input1" help:"Executes a query when sqlcmd starts, but does not exit sqlcmd when the query has finished running. Multiple-semicolon-delimited queries can be executed."`
	// Query to run then exit
	Query  string `short:"Q" xor:"input2" help:"Executes a query when sqlcmd starts and then immediately exits sqlcmd. Multiple-semicolon-delimited queries can be executed."`
	Server string `short:"S" help:"[tcp:]server[\\instance_name][,port]Specifies the instance of SQL Server to which to connect. It sets the sqlcmd scripting variable SQLCMDSERVER."`
	// Disable syscommands with a warning
	DisableCmdAndWarn bool `short:"X" xor:"syscmd" help:"Disables commands that might compromise system security. Sqlcmd issues a warning and continues."`
}
type logger struct {
	s *sqlcmd.Sqlcmd
}

func (l logger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(l.s.GetOutput(), format, v...)
}

func (l logger) Println(v ...interface{}) {
	fmt.Fprintln(l.s.GetOutput(), v...)
}

var Args SqlCmdArguments

func main() {
	kong.Parse(&Args)
	vars := variables.InitializeVariables(!Args.DisableCmdAndWarn)
	setVars(vars, &Args)

	exitCode, err := run(vars)
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}

// Initializes scripting variables from command line arguments
func setVars(vars *variables.Variables, args *SqlCmdArguments) {
	varmap := map[string]func(*SqlCmdArguments) string{
		variables.SQLCMDDBNAME:            func(a *SqlCmdArguments) string { return a.DatabaseName },
		variables.SQLCMDLOGINTIMEOUT:      func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDUSEAAD:            func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDWORKSTATION:       func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDPASSWORD:          func(a *SqlCmdArguments) string { return a.Password },
		variables.SQLCMDSERVER:            func(a *SqlCmdArguments) string { return a.Server },
		variables.SQLCMDERRORLEVEL:        func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDPACKETSIZE:        func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDUSER:              func(a *SqlCmdArguments) string { return a.UserName },
		variables.SQLCMDSTATTIMEOUT:       func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDHEADERS:           func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDCOLSEP:            func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDCOLDWIDTH:         func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDMAXVARTYPEWIDTH:   func(a *SqlCmdArguments) string { return "" },
		variables.SQLCMDMAXFIXEDTYPEWIDTH: func(a *SqlCmdArguments) string { return "" },
	}
	for varname, set := range varmap {
		val := set(args)
		if val != "" {
			vars.Set(varname, val)
		}
	}
}

func run(vars *variables.Variables) (exitcode int, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	if Args.BatchTerminator != "GO" {
		err = sqlcmd.SetBatchTerminator(Args.BatchTerminator)
		if err != nil {
			err = fmt.Errorf("invalid batch terminator '%s'", Args.BatchTerminator)
		}
	}
	if err != nil {
		return 1, err
	}

	iactive := Args.Query == "" && Args.InputFile == nil
	line, err := rline.New(!iactive, "", "")
	if err != nil {
		return 1, err
	}
	defer line.Close()
	s := sqlcmd.New(line, wd, vars)
	s.Connect.UseTrustedConnection = Args.UseTrustedConnection
	s.Connect.TrustServerCertificate = Args.TrustServerCertificate
	s.Format = sqlcmd.NewSqlCmdDefaultFormatter(false)
	if Args.OutputFile != "" {
		err = sqlcmd.Out(s, []string{Args.OutputFile}, 0)
		if err != nil {
			return 1, err
		}
	}
	err = s.ConnectDb("", "", "", !iactive)
	if err != nil {
		return 1, err
	}
	if iactive {
		err = s.Run()
	}
	return s.Exitcode, err
}
