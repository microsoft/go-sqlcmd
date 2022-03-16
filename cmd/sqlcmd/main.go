// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/denisenkom/go-mssqldb/azuread"
	"github.com/microsoft/go-sqlcmd/pkg/console"
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
	PacketSize                  int               `short:"a" help:"Requests a packet of a different size. This option sets the sqlcmd scripting variable SQLCMDPACKETSIZE. packet_size must be a value between 512 and 32767. The default = 4096. A larger packet size can enhance performance for execution of scripts that have lots of SQL statements between GO commands. You can request a larger packet size. However, if the request is denied, sqlcmd uses the server default for packet size."`
	LoginTimeout                int               `short:"l" default:"-1" help:"Specifies the number of seconds before a sqlcmd login to the go-mssqldb driver times out when you try to connect to a server. This option sets the sqlcmd scripting variable SQLCMDLOGINTIMEOUT. The default value is 30. 0 means infinite."`
	WorkstationName             string            `short:"H" help:"This option sets the sqlcmd scripting variable SQLCMDWORKSTATION. The workstation name is listed in the hostname column of the sys.sysprocesses catalog view and can be returned using the stored procedure sp_who. If this option is not specified, the default is the current computer name. This name can be used to identify different sqlcmd sessions."`
	ApplicationIntent           string            `short:"K" default:"default" enum:"default,ReadOnly" help:"Declares the application workload type when connecting to a server. The only currently supported value is ReadOnly. If -K is not specified, the sqlcmd utility will not support connectivity to a secondary replica in an Always On availability group."`
	EncryptConnection           string            `short:"N" default:"default" enum:"default,false,true,disable" help:"This switch is used by the client to request an encrypted connection."`
	DriverLoggingLevel          int               `help:"Level of mssql driver messages to print."`
	ExitOnError                 bool              `short:"b" help:"Specifies that sqlcmd exits and returns a DOS ERRORLEVEL value when an error occurs."`
	ErrorSeverityLevel          uint8             `short:"V" help:"Controls the severity level that is used to set the ERRORLEVEL variable on exit."`
	ErrorLevel                  int               `short:"m" help:"Controls which error messages are sent to stdout. Messages that have severity level greater than or equal to this level are sent."`
	Format                      string            `short:"F" help:"Specifies the formatting for results." default:"horiz" enum:"horiz,horizontal,vert,vertical"`
	ErrorsToStderr              int               `short:"r" help:"Redirects the error message output to the screen (stderr). A value of 0 means messages with severity >= 11 will b redirected. A value of 1 means all error message output including PRINT is redirected." enum:"-1,0,1" default:"-1"`
}

// Validate accounts for settings not described by Kong attributes
func (a *SQLCmdArguments) Validate() error {
	if a.PacketSize != 0 && (a.PacketSize < 512 || a.PacketSize > 32767) {
		return fmt.Errorf(`'-a %d': Packet size has to be a number between 512 and 32767.`, a.PacketSize)
	}

	return nil
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
			return azuread.ActiveDirectoryIntegrated
		case hasPassword:
			return azuread.ActiveDirectoryPassword
		default:
			return azuread.ActiveDirectoryInteractive
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
	exitCode, _ := run(vars, &args)
	os.Exit(exitCode)
}

// setVars initializes scripting variables from command line arguments
func setVars(vars *sqlcmd.Variables, args *SQLCmdArguments) {
	varmap := map[string]func(*SQLCmdArguments) string{
		sqlcmd.SQLCMDDBNAME: func(a *SQLCmdArguments) string { return a.DatabaseName },
		sqlcmd.SQLCMDLOGINTIMEOUT: func(a *SQLCmdArguments) string {
			if a.LoginTimeout > -1 {
				return fmt.Sprint(a.LoginTimeout)
			}
			return ""
		},
		sqlcmd.SQLCMDUSEAAD: func(a *SQLCmdArguments) string {
			if a.UseAad {
				return "true"
			}
			switch a.AuthenticationMethod {
			case azuread.ActiveDirectoryIntegrated:
			case azuread.ActiveDirectoryInteractive:
			case azuread.ActiveDirectoryPassword:
				return "true"
			}
			return ""
		},
		sqlcmd.SQLCMDWORKSTATION: func(a *SQLCmdArguments) string { return args.WorkstationName },
		sqlcmd.SQLCMDSERVER:      func(a *SQLCmdArguments) string { return a.Server },
		sqlcmd.SQLCMDERRORLEVEL:  func(a *SQLCmdArguments) string { return fmt.Sprint(a.ErrorLevel) },
		sqlcmd.SQLCMDPACKETSIZE: func(a *SQLCmdArguments) string {
			if args.PacketSize > 0 {
				return fmt.Sprint(args.PacketSize)
			}
			return ""
		},
		sqlcmd.SQLCMDUSER:              func(a *SQLCmdArguments) string { return a.UserName },
		sqlcmd.SQLCMDSTATTIMEOUT:       func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDHEADERS:           func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDCOLSEP:            func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDCOLWIDTH:          func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDMAXVARTYPEWIDTH:   func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDMAXFIXEDTYPEWIDTH: func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDFORMAT:            func(a *SQLCmdArguments) string { return a.Format },
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

func setConnect(connect *sqlcmd.ConnectSettings, args *SQLCmdArguments, vars *sqlcmd.Variables) {
	if !args.DisableCmdAndWarn {
		connect.Password = os.Getenv(sqlcmd.SQLCMDPASSWORD)
	}
	connect.ServerName = args.Server
	if connect.ServerName == "" {
		connect.ServerName, _ = vars.Get(sqlcmd.SQLCMDSERVER)
	}
	connect.Database = args.DatabaseName
	if connect.Database == "" {
		connect.Database, _ = vars.Get(sqlcmd.SQLCMDDBNAME)
	}
	connect.UserName = args.UserName
	if connect.UserName == "" {
		connect.UserName, _ = vars.Get(sqlcmd.SQLCMDUSER)
	}
	connect.UseTrustedConnection = args.UseTrustedConnection
	connect.TrustServerCertificate = args.TrustServerCertificate
	connect.AuthenticationMethod = args.authenticationMethod(connect.Password != "")
	connect.DisableEnvironmentVariables = args.DisableCmdAndWarn
	connect.DisableVariableSubstitution = args.DisableVariableSubstitution
	connect.ApplicationIntent = args.ApplicationIntent
	connect.LoginTimeoutSeconds = args.LoginTimeout
	connect.Encrypt = args.EncryptConnection
	connect.PacketSize = args.PacketSize
	connect.WorkstationName = args.WorkstationName
	connect.LogLevel = args.DriverLoggingLevel
	connect.ExitOnError = args.ExitOnError
	connect.ErrorSeverityLevel = args.ErrorSeverityLevel
}

func run(vars *sqlcmd.Variables, args *SQLCmdArguments) (int, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	iactive := args.InputFile == nil && args.Query == ""
	var line sqlcmd.Console = nil
	if iactive {
		line = console.NewConsole("")
	}

	s := sqlcmd.New(line, wd, vars)
	setConnect(&s.Connect, args, vars)
	if args.BatchTerminator != "GO" {
		err = s.Cmd.SetBatchTerminator(args.BatchTerminator)
		if err != nil {
			err = fmt.Errorf("invalid batch terminator '%s'", args.BatchTerminator)
		}
	}
	if err != nil {
		return 1, err
	}

	setConnect(&s.Connect, args, vars)
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(false)
	if args.OutputFile != "" {
		err = s.RunCommand(s.Cmd["OUT"], []string{args.OutputFile})
		if err != nil {
			return 1, err
		}
	} else {
		var stderrSeverity uint8 = 11
		if args.ErrorsToStderr == 1 {
			stderrSeverity = 0
		}
		if args.ErrorsToStderr >= 0 {
			s.PrintError = func(msg string, severity uint8) bool {
				if severity >= stderrSeverity {
					_, _ = os.Stderr.Write([]byte(msg))
					return true
				}
				return false
			}
		}
	}
	once := false
	if args.InitialQuery != "" {
		s.Query = args.InitialQuery
	} else if args.Query != "" {
		once = true
		s.Query = args.Query
	}
	// connect using no overrides
	err = s.ConnectDb(nil, !iactive)
	if err != nil {
		return 1, err
	}
	if iactive || s.Query != "" {
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
