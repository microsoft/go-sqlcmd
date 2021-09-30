// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/gohxs/readline"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

// SQLCmdArguments defines the command line arguments for sqlcmd
// The exhaustive list is at https://docs.microsoft.com/sql/tools/sqlcmd-utility?view=sql-server-ver15
type SQLCmdArguments struct {
	// Which batch terminator to use. Default is GO
	BatchTerminator string `short:"c" default:"GO" arghelp:"Specifies the batch terminator. The default value is GO."`
	// Whether to trust the server certificate on an encrypted connection
	TrustServerCertificate bool   `short:"C" help:"Implicitly trust the server certificate without validation."`
	DatabaseName           string `short:"d" help:"This option sets the sqlcmd scripting variable SQLCMDDBNAME. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits."`
	UseTrustedConnection   bool   `short:"E" xor:"uid, auth" help:"Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any any environment variables that define user name and password."`
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
	// AuthenticationMethod is new for go-sqlcmd
	AuthenticationMethod        string            `xor:"auth"  help:"Specifies the SQL authentication method to use to connect to Azure SQL Database. One of:ActiveDirectoryDefault,ActiveDirectoryIntegrated,ActiveDirectoryPassword,ActiveDirectoryInteractive,ActiveDirectoryManagedIdentity,ActiveDirectoryServicePrincipal,SqlPassword"`
	UseAad                      bool              `short:"G" xor:"auth" help:"Tells sqlcmd to use Active Directory authentication. If no user name is provided, authentication method ActiveDirectoryDefault is used. If a password is provided, ActiveDirectoryPassword is used. Otherwise ActiveDirectoryInteractive is used."`
	DisableVariableSubstitution bool              `short:"x" help:"Causes sqlcmd to ignore scripting variables. This parameter is useful when a script contains many INSERT statements that may contain strings that have the same format as regular variables, such as $(variable_name)."`
	Variables                   map[string]string `short:"v" help:"Creates a sqlcmd scripting variable that can be used in a sqlcmd script. Enclose the value in quotation marks if the value contains spaces. You can specify multiple var=values values. If there are errors in any of the values specified, sqlcmd generates an error message and then exits"`
}

// newArguments constructs a SQLCmdArguments instance with default values
// Any parameter with a "default" Kong attribute should have an assignment here
func newArguments() SQLCmdArguments {
	return SQLCmdArguments{
		BatchTerminator: "GO",
	}
}

// Breaking changes in command line are listed here.
// Any switch not listed in breaking changes and not also included in SqlCmdArguments just has not been implemented yet
// 1. -P: Passwords have to be provided through SQLCMDPASSWORD environment variable or typed when prompted
// 2. -R: Go runtime doesn't expose user locale information and syscall would only enable it on Windows, so we won't try to implement it

var args SQLCmdArguments

func (a SQLCmdArguments) authenticationMethod(hasPassword bool) string {
	if a.UseTrustedConnection {
		return sqlcmd.NotSpecified
	}
	if a.UseAad {
		switch {
		case a.UserName == "":
			return sqlcmd.ActiveDirectoryIntegrated
		case hasPassword:
			return sqlcmd.ActiveDirectoryPassword
		default:
			return sqlcmd.ActiveDirectoryInteractive
		}
	}
	if a.AuthenticationMethod == "" {
		return sqlcmd.NotSpecified
	}
	return a.AuthenticationMethod
}

func main() {
	kong.Parse(&args)
	vars := sqlcmd.InitializeVariables(!args.DisableCmdAndWarn)
	setVars(vars, &args)

	// so far sqlcmd prints all the errors itself so ignore it
	exitCode, _ := run(vars)
	os.Exit(exitCode)
}

// setVars initializes scripting variables from command line arguments
func setVars(vars *sqlcmd.Variables, args *SQLCmdArguments) {
	varmap := map[string]func(*SQLCmdArguments) string{
		sqlcmd.SQLCMDDBNAME:       func(a *SQLCmdArguments) string { return a.DatabaseName },
		sqlcmd.SQLCMDLOGINTIMEOUT: func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDUSEAAD: func(a *SQLCmdArguments) string {
			if a.UseAad {
				return "true"
			}
			switch a.AuthenticationMethod {
			case sqlcmd.ActiveDirectoryIntegrated:
			case sqlcmd.ActiveDirectoryInteractive:
			case sqlcmd.ActiveDirectoryPassword:
				return "true"
			}
			return ""
		},
		sqlcmd.SQLCMDWORKSTATION:       func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDSERVER:            func(a *SQLCmdArguments) string { return a.Server },
		sqlcmd.SQLCMDERRORLEVEL:        func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDPACKETSIZE:        func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDUSER:              func(a *SQLCmdArguments) string { return a.UserName },
		sqlcmd.SQLCMDSTATTIMEOUT:       func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDHEADERS:           func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDCOLSEP:            func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDCOLWIDTH:          func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDMAXVARTYPEWIDTH:   func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDMAXFIXEDTYPEWIDTH: func(a *SQLCmdArguments) string { return "" },
	}
	for varname, set := range varmap {
		val := set(args)
		if val != "" {
			vars.Set(varname, val)
		}
	}

	// Following sqlcmd tradition there's no validation of -v kvps
	for v := range args.Variables {
		vars.Set(v, args.Variables[v])
	}

}

func run(vars *sqlcmd.Variables) (int, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	iactive := args.InputFile == nil
	var line *readline.Instance
	if iactive {
		line, err = readline.New(">")
		if err != nil {
			return 1, err
		}
		defer line.Close()
	}

	s := sqlcmd.New(line, wd, vars)
	if !args.DisableCmdAndWarn {
		s.Connect.Password = os.Getenv(sqlcmd.SQLCMDPASSWORD)
	}
	if args.BatchTerminator != "GO" {
		err = s.Cmd.SetBatchTerminator(args.BatchTerminator)
		if err != nil {
			err = fmt.Errorf("invalid batch terminator '%s'", args.BatchTerminator)
		}
	}
	if err != nil {
		return 1, err
	}
	s.Connect.UseTrustedConnection = args.UseTrustedConnection
	s.Connect.TrustServerCertificate = args.TrustServerCertificate
	s.Connect.AuthenticationMethod = args.authenticationMethod(s.Connect.Password != "")
	s.Connect.DisableEnvironmentVariables = args.DisableCmdAndWarn
	s.Connect.DisableVariableSubstitution = args.DisableVariableSubstitution
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(false)
	if args.OutputFile != "" {
		err = s.RunCommand(s.Cmd["OUT"], []string{args.OutputFile})
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
		err = s.Run(once, false)
	} else {
		for f := range args.InputFile {
			if err = s.IncludeFile(args.InputFile[f], true); err != nil {
				break
			}
		}
	}
	s.SetOutput(nil)
	s.SetError(nil)
	return s.Exitcode, err
}
