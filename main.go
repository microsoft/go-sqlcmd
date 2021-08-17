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

// SQLCmdArguments defines the command line arguments for sqlcmd
// The exhaustive list is at https://docs.microsoft.com/sql/tools/sqlcmd-utility?view=sql-server-ver15
type SQLCmdArguments struct {
	// Which batch terminator to use. Default is GO
	BatchTerminator string `short:"c" default:"GO" arghelp:"Specifies the batch terminator. The default value is GO."`
	// Whether to trust the server certificate on an encrypted connection
	TrustServerCertificate bool   `short:"C" help:"Implicitly trust the server certificate without validation."`
	DatabaseName           string `short:"d" help:"This option sets the sqlcmd scripting variable SQLCMDDBNAME. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits."`
	UseTrustedConnection   bool   `short:"E" xor:"uid" help:"Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any any environment variables that define user name and password."`
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

// Breaking changes in command line are listed here.
// Any switch not listed in breaking changes and not also included in SqlCmdArguments just has not been implemented yet
// 1. -P: Passwords have to be provided through SQLCMDPASSWORD environment variable or typed when prompted
// 2. -R: Go runtime doesn't expose user locale information and syscall would only enable it on Windows, so we won't try to implement it

var args SQLCmdArguments

func main() {
	kong.Parse(&args)
	vars := variables.InitializeVariables(!args.DisableCmdAndWarn)
	setVars(vars, &args)

	exitCode, err := run(vars)
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}

// Initializes scripting variables from command line arguments
func setVars(vars *variables.Variables, args *SQLCmdArguments) {
	varmap := map[string]func(*SQLCmdArguments) string{
		variables.SQLCMDDBNAME:            func(a *SQLCmdArguments) string { return a.DatabaseName },
		variables.SQLCMDLOGINTIMEOUT:      func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDUSEAAD:            func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDWORKSTATION:       func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDSERVER:            func(a *SQLCmdArguments) string { return a.Server },
		variables.SQLCMDERRORLEVEL:        func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDPACKETSIZE:        func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDUSER:              func(a *SQLCmdArguments) string { return a.UserName },
		variables.SQLCMDSTATTIMEOUT:       func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDHEADERS:           func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDCOLSEP:            func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDCOLWIDTH:          func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDMAXVARTYPEWIDTH:   func(a *SQLCmdArguments) string { return "" },
		variables.SQLCMDMAXFIXEDTYPEWIDTH: func(a *SQLCmdArguments) string { return "" },
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
	if args.BatchTerminator != "GO" {
		err = sqlcmd.SetBatchTerminator(args.BatchTerminator)
		if err != nil {
			err = fmt.Errorf("invalid batch terminator '%s'", args.BatchTerminator)
		}
	}
	if err != nil {
		return 1, err
	}

	iactive := args.InputFile == nil
	line, err := rline.New(!iactive, "", "")
	if err != nil {
		return 1, err
	}
	defer line.Close()

	s := sqlcmd.New(line, wd, vars)
	s.Connect.UseTrustedConnection = args.UseTrustedConnection
	s.Connect.TrustServerCertificate = args.TrustServerCertificate
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(false)
	if args.OutputFile != "" {
		err = sqlcmd.Out(s, []string{args.OutputFile}, 0)
		if err != nil {
			return 1, err
		}
	}
	once := false
	if args.InitialQuery != "" {
		s.Query = args.InitialQuery
	} else if args.Query != "" {
		once = true
		s.Query = args.Query
	}
	err = s.ConnectDb("", "", "", !iactive)
	if err != nil {
		return 1, err
	}
	if iactive {
		err = s.Run(once)
	}
	return s.Exitcode, err
}
