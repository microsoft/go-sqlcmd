// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
//

package sqlcmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/pkg/console"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SQLCmdArguments defines the command line arguments for sqlcmd
// The exhaustive list is at https://docs.microsoft.com/sql/tools/sqlcmd-utility?view=sql-server-ver15
type SQLCmdArguments struct {
	// Which batch terminator to use. Default is GO
	BatchTerminator string
	// Whether to trust the server certificate on an encrypted connection
	TrustServerCertificate bool
	DatabaseName           string
	UseTrustedConnection   bool
	UserName               string
	// Files from which to read query text
	InputFile  []string
	OutputFile string
	// First query to run in interactive mode
	InitialQuery string
	// Query to run then exit
	Query  string
	Server string
	// Disable syscommands with a warning
	DisableCmdAndWarn bool
	// AuthenticationMethod is new for go-sqlcmd
	AuthenticationMethod        string
	UseAad                      bool
	DisableVariableSubstitution bool
	Variables                   map[string]string
	PacketSize                  int
	LoginTimeout                int
	WorkstationName             string
	ApplicationIntent           string
	EncryptConnection           string
	DriverLoggingLevel          int
	ExitOnError                 bool
	ErrorSeverityLevel          uint8
	ErrorLevel                  int
	Format                      string
	ErrorsToStderr              int
	Headers                     int
	UnicodeOutputFile           bool
	Version                     bool
	ColumnSeparator             string
	ScreenWidth                 *int
	TrimSpaces                  bool
	MultiSubnetFailover         bool
	Password                    string
	// Keep Help at the end of the list
	Help bool
}

// Validate arguments for settings not described by Kong attributes
func (a *SQLCmdArguments) Validate() error {
	if a.PacketSize != 0 && (a.PacketSize < 512 || a.PacketSize > 32767) {
		return localizer.Errorf(`'-a %d': Packet size has to be a number between 512 and 32767.`, a.PacketSize)
	}
	// Ignore 0 even though it's technically an invalid input
	if a.Headers < -1 {
		return localizer.Errorf(`'-h %d': header value must be either -1 or a value between 1 and 2147483647`, a.Headers)
	}
	if a.ScreenWidth != nil && (*a.ScreenWidth < 9 || *a.ScreenWidth > 65535) {
		return localizer.Errorf(`'-w %d': value must be greater than 8 and less than 65536.`, *a.ScreenWidth)
	}
	if a.Password != "" {
		return localizer.Errorf(`'-P' is obsolete. The initial passwords must be set using the SQLCMDPASSWORD environment variable or entered at the password prompt.`)
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

func Execute(version string) {
	//var screenWidth *int
	rootCmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, argss []string) error {
			SetScreenWidthFlag(&args, cmd)
			if err := args.Validate(); err != nil {
				cmd.SilenceUsage = true
				return err
			}
			if err := normalizeFlags(cmd); err != nil {
				cmd.SilenceUsage = true
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, argss []string) {
			if len(argss) > 0 {
				fmt.Printf("Sqlcmd: '%s': Unknown command. Enter '--help' for command help.", argss[0])
				os.Exit(1)
			}

			vars := sqlcmd.InitializeVariables(!args.DisableCmdAndWarn)
			setVars(vars, &args)

			if args.Version {
				fmt.Printf("%v\n", version)
				fmt.Println(localizer.Sprintf("Legal docs and information: aka.ms/SqlcmdLegal"))
				fmt.Println(localizer.Sprintf("Third party notices: aka.ms/SqlcmdNotices"))
				os.Exit(0)
			}

			if args.Help {
				fmt.Print(cmd.UsageString())
				os.Exit(0)
			}

			exitCode, _ := run(vars, &args)
			os.Exit(exitCode)

		},
	}
	setFlags(rootCmd, &args)
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, argss []string) {
		fmt.Println(cmd.Long)
		cmd.Flags().SetInterspersed(false)
		fmt.Println("Flags:")
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if len(flag.Shorthand) > 0 {
				fmt.Printf("  -%s, --%s\n", flag.Shorthand, flag.Name)
			} else {
				fmt.Printf("      --%s\n", flag.Name)
			}
			desc := formatDescription(flag.Usage, 60, 8)
			fmt.Printf("        %s\n", desc)
			fmt.Println()
		})
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func formatDescription(description string, maxWidth, indentWidth int) string {
	var lines []string
	words := strings.Fields(description)
	line := ""
	for _, word := range words {
		if len(line)+len(word)+1 <= maxWidth {
			line += word + " "
		} else {
			lines = append(lines, line)
			line = strings.Repeat(" ", indentWidth) + word + " "
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func SetScreenWidthFlag(args *SQLCmdArguments, rootCmd *cobra.Command) {
	screenWidth := rootCmd.Flags().Lookup("screen-width")
	if screenWidth == nil {
		return
	}
	value := screenWidth.Value.String()
	if value != screenWidth.DefValue {
		args.ScreenWidth = new(int)

		screenWidthValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("Error converting screen width:", err)
			return
		}
		args.ScreenWidth = &screenWidthValue
	} else {
		args.ScreenWidth = nil
	}
}

func setFlags(rootCmd *cobra.Command, args *SQLCmdArguments) {
	rootCmd.Flags().BoolVarP(&args.Help, "help", "?", false, "-? shows this syntax summary, --help shows modern sqlcmd sub-command help")
	var inputfiles []string
	rootCmd.Flags().StringSliceVarP(&args.InputFile, "input-file", "i", inputfiles, "Identifies one or more files that contain batches of SQL statements. If one or more files do not exist, sqlcmd will exit. Mutually exclusive with -Q/-q.")
	rootCmd.Flags().StringVarP(&args.OutputFile, "output-file", "o", "", "Identifies the file that receives output from sqlcmd.")
	rootCmd.Flags().BoolVarP(&args.Version, "version", "", false, "Print version information and exit")
	rootCmd.Flags().BoolVarP(&args.TrustServerCertificate, "trust-server-certificate", "C", false, "Implicitly trust the server certificate without validation.")
	rootCmd.Flags().StringVarP(&args.DatabaseName, "database-name", "d", "", "This option sets the sqlcmd scripting variable SQLCMDDBNAME. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits.")
	rootCmd.Flags().BoolVarP(&args.UseTrustedConnection, "use-trusted-connection", "E", false, "Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any environment variables that define user name and password.")
	rootCmd.Flags().StringVarP(&args.BatchTerminator, "batch-terminator", "c", "GO", "Specifies the batch terminator. The default value is GO")
	rootCmd.Flags().StringVarP(&args.UserName, "user-name", "U", "", "The login name or contained database user name.  For contained database users, you must provide the database name option")
	rootCmd.Flags().StringVarP(&args.InitialQuery, "initial-query", "q", "", "Executes a query when sqlcmd starts, but does not exit sqlcmd when the query has finished running. Multiple-semicolon-delimited queries can be executed.")
	rootCmd.Flags().StringVarP(&args.Query, "query", "Q", "", "Executes a query when sqlcmd starts and then immediately exits sqlcmd. Multiple-semicolon-delimited queries can be executed.")
	rootCmd.Flags().StringVarP(&args.Server, "server", "S", "", "[[tcp:]|[lpc:]|[np:]]server[\\instance_name][,port]Specifies the instance of SQL Server to which to connect. It sets the sqlcmd scripting variable SQLCMDSERVER.")
	rootCmd.Flags().BoolVarP(&args.DisableCmdAndWarn, "disable-cmd-and-warn", "X", false, "Disables commands that might compromise system security. Sqlcmd issues a warning and continues.")
	rootCmd.Flags().StringVar(&args.AuthenticationMethod, "authentication-method", "", "Specifies the SQL authentication method to use to connect to Azure SQL Database. One of: ActiveDirectoryDefault, ActiveDirectoryIntegrated, ActiveDirectoryPassword, ActiveDirectoryInteractive, ActiveDirectoryManagedIdentity, ActiveDirectoryServicePrincipal, SqlPassword")
	rootCmd.Flags().BoolVarP(&args.UseAad, "use-aad", "G", false, "Tells sqlcmd to use Active Directory authentication. If no user name is provided, authentication method ActiveDirectoryDefault is used. If a password is provided, ActiveDirectoryPassword is used. Otherwise ActiveDirectoryInteractive is used.")
	rootCmd.Flags().BoolVarP(&args.DisableVariableSubstitution, "disable-variable-substitution", "x", false, "Causes sqlcmd to ignore scripting variables. This parameter is useful when a script contains many INSERT statements that may contain strings that have the same format as regular variables, such as $(variable_name).")
	var variables map[string]string
	rootCmd.Flags().StringToStringVarP(&args.Variables, "variables", "v", variables, "Creates a sqlcmd scripting variable that can be used in a sqlcmd script. Enclose the value in quotation marks if the value contains spaces. You can specify multiple var=values values. If there are errors in any of the values specified, sqlcmd generates an error message and then exits")
	rootCmd.Flags().IntVarP(&args.PacketSize, "packet-size", "a", 0, "Requests a packet of a different size. This option sets the sqlcmd scripting variable SQLCMDPACKETSIZE. packet_size must be a value between 512 and 32767. The default = 4096. A larger packet size can enhance performance for execution of scripts that have lots of SQL statements between GO commands. You can request a larger packet size. However, if the request is denied, sqlcmd uses the server default for packet size.")
	rootCmd.Flags().IntVarP(&args.LoginTimeout, "login-timeOut", "l", -1, "Specifies the number of seconds before a sqlcmd login to the go-mssqldb driver times out when you try to connect to a server. This option sets the sqlcmd scripting variable SQLCMDLOGINTIMEOUT. The default value is 30. 0 means infinite.")
	rootCmd.Flags().StringVarP(&args.WorkstationName, "workstation-name", "H", "", "This option sets the sqlcmd scripting variable SQLCMDWORKSTATION. The workstation name is listed in the hostname column of the sys.sysprocesses catalog view and can be returned using the stored procedure sp_who. If this option is not specified, the default is the current computer name. This name can be used to identify different sqlcmd sessions.")

	rootCmd.Flags().StringVarP(&args.ApplicationIntent, "application-intent", "K", "default", "Declares the application workload type when connecting to a server. The only currently supported value is ReadOnly. If -K is not specified, the sqlcmd utility will not support connectivity to a secondary replica in an Always On availability group")
	rootCmd.Flags().StringVarP(&args.EncryptConnection, "encrypt-connection", "N", "default", "This switch is used by the client to request an encrypted connection.")
	rootCmd.Flags().StringVarP(&args.Format, "format", "F", "horiz", "Specifies the formatting for results.")
	rootCmd.Flags().IntVarP(&args.ErrorsToStderr, "errors-to-stderr", "r", -1, "Controls which error messages are sent to stdout. Messages that have severity level greater than or equal to this level are sent.")

	rootCmd.Flags().IntVar(&args.DriverLoggingLevel, "driver-logging-level", 0, "Level of mssql driver messages to print.")
	rootCmd.Flags().BoolVarP(&args.ExitOnError, "exit-on-error", "b", false, "Specifies that sqlcmd exits and returns a DOS ERRORLEVEL value when an error occurs.")
	rootCmd.Flags().IntVarP(&args.ErrorLevel, "error-level", "m", 0, "Controls which error messages are sent to stdout. Messages that have severity level greater than or equal to this level are sent.")

	//Need to decide on short of Header , as "h" is already used in help command in Cobra
	rootCmd.Flags().IntVarP(&args.Headers, "headers", "h", 0, "Specifies the number of rows to print between the column headings. Use -h-1 to specify that headers not be printed.")

	rootCmd.Flags().BoolVarP(&args.UnicodeOutputFile, "unicode-output-file", "u", false, "Specifies that all output files are encoded with little-endian Unicode")
	rootCmd.Flags().StringVarP(&args.ColumnSeparator, "column-separator", "s", "", "Specifies the column separator character. Sets the SQLCMDCOLSEP variable.")
	rootCmd.Flags().BoolVarP(&args.TrimSpaces, "trim-spaces", "W", false, "Remove trailing spaces from a column.")
	rootCmd.Flags().BoolVarP(&args.MultiSubnetFailover, "multi-subnet-failover", "M", false, "Provided for backward compatibility. Sqlcmd always optimizes detection of the active replica of a SQL Failover Cluster.")

	rootCmd.Flags().StringVarP(&args.Password, "password", "P", "", "Obsolete. The initial passwords must be set using the SQLCMDPASSWORD environment variable or entered at the password prompt.")

	// Using PersistentFlags() for ErrorSeverityLevel due to data type uint8 , which is not supported in Flags()
	rootCmd.PersistentFlags().Uint8VarP(&args.ErrorSeverityLevel, "error-severity-level", "V", 0, "Controls the severity level that is used to set the ERRORLEVEL variable on exit.")

	var screenWidth int
	screenWidthPtr := &screenWidth
	rootCmd.Flags().IntVarP(screenWidthPtr, "screen-width", "w", 0, localizer.Sprintf("Specifies the screen width for output"))
}

func normalizeFlags(rootCmd *cobra.Command) error {
	//Adding a validator for checking the enum flags
	var err error
	rootCmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "application-intent":
			value := strings.ToLower(getFlagValueByName(f, name))
			switch value {
			case "default", "readonly":
				return pflag.NormalizedName(name)
			default:
				err = fmt.Errorf("--application-intent must be one of \"default\",\"readonly\" but got \"%s\"", value)
				return pflag.NormalizedName("")
			}
		case "encrypt-connection":
			value := strings.ToLower(getFlagValueByName(f, name))
			switch value {
			case "default", "false", "true", "disable":
				return pflag.NormalizedName(name)
			default:
				err = fmt.Errorf("--encrypt-connection must be one of \"default\",\"false\",\"true\",\"disable\" but got \"%s\"", value)
				return pflag.NormalizedName("")
			}
		case "format":
			value := strings.ToLower(getFlagValueByName(f, name))
			switch value {
			case "horiz", "horizontal", "vert", "vertical":
				return pflag.NormalizedName(name)
			default:
				err = fmt.Errorf("--format must be one of \"horiz\",\"horizontal\",\"vert\",\"vertical\" but got \"%s\"", value)
				return pflag.NormalizedName("")
			}
		case "errors-to-stderr":
			value := getFlagValueByName(f, name)
			switch value {
			case "-1", "0", "1":
				return pflag.NormalizedName(name)
			default:
				err = fmt.Errorf("--errors-to-stderr must be one of \"-1\",\"0\",\"1\" but got \"%s\"", value)
				return pflag.NormalizedName("")
			}
		}
		return pflag.NormalizedName(name)
	})
	return err
}
func getFlagValueByName(flagSet *pflag.FlagSet, name string) string {
	var value string
	flagSet.VisitAll(func(f *pflag.Flag) {
		if f.Name == name {
			value = f.Value.String()
			return
		}
	})
	return value
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
		sqlcmd.SQLCMDUSER:        func(a *SQLCmdArguments) string { return a.UserName },
		sqlcmd.SQLCMDSTATTIMEOUT: func(a *SQLCmdArguments) string { return "" },
		sqlcmd.SQLCMDHEADERS:     func(a *SQLCmdArguments) string { return fmt.Sprint(a.Headers) },
		sqlcmd.SQLCMDCOLSEP: func(a *SQLCmdArguments) string {
			if a.ColumnSeparator != "" {
				return string(a.ColumnSeparator[0])
			}
			return ""
		},
		sqlcmd.SQLCMDCOLWIDTH: func(a *SQLCmdArguments) string {
			if a.ScreenWidth != nil {
				return fmt.Sprint(*a.ScreenWidth)
			}
			return ""
		},
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
	connect.ApplicationName = "sqlcmd"
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

func isConsoleInitializationRequired(connect *sqlcmd.ConnectSettings, args *SQLCmdArguments) bool {
	iactive := args.InputFile == nil && args.Query == ""
	return iactive || connect.RequiresPassword()
}

func run(vars *sqlcmd.Variables, args *SQLCmdArguments) (int, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	var connectConfig sqlcmd.ConnectSettings
	setConnect(&connectConfig, args, vars)
	var line sqlcmd.Console = nil
	if isConsoleInitializationRequired(&connectConfig, args) {
		line = console.NewConsole("")
		defer line.Close()
	}

	s := sqlcmd.New(line, wd, vars)
	// We want the default behavior on ctrl-c - exit the process
	s.SetupCloseHandler()
	defer s.StopCloseHandler()
	s.UnicodeOutputFile = args.UnicodeOutputFile

	if args.DisableCmdAndWarn {
		s.Cmd.DisableSysCommands(false)
	}

	if args.BatchTerminator != "GO" {
		err = s.Cmd.SetBatchTerminator(args.BatchTerminator)
		if err != nil {
			err = localizer.Errorf("invalid batch terminator '%s'", args.BatchTerminator)
		}
	}
	if err != nil {
		return 1, err
	}

	s.Connect = &connectConfig
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(args.TrimSpaces)
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
					s.WriteError(os.Stderr, errors.New(msg+sqlcmd.SqlcmdEol))
					return true
				}
				return false
			}
		}
	}

	// connect using no overrides
	err = s.ConnectDb(nil, line == nil)
	if err != nil {
		s.WriteError(s.GetError(), err)
		return 1, err
	}

	script := vars.StartupScriptFile()
	if !args.DisableCmdAndWarn && len(script) > 0 {
		f, fileErr := os.Open(script)
		if fileErr != nil {
			s.WriteError(s.GetError(), sqlcmd.InvalidVariableValue(sqlcmd.SQLCMDINI, script))
		} else {
			_ = f.Close()
			// IncludeFile won't return an error for a SQL error, but ExitCode will be non-zero if -b was passed on the commandline
			err = s.IncludeFile(script, true)
		}
	}

	if err == nil && s.Exitcode == 0 {
		once := false
		if args.InitialQuery != "" {
			s.Query = args.InitialQuery
		} else if args.Query != "" {
			once = true
			s.Query = args.Query
		}
		iactive := args.InputFile == nil && args.Query == ""
		if iactive || s.Query != "" {
			err = s.Run(once, false)
		} else {
			for f := range args.InputFile {
				if err = s.IncludeFile(args.InputFile[f], true); err != nil {
					s.WriteError(s.GetError(), err)
					s.Exitcode = 1
					break
				}
			}
		}
	}
	s.SetOutput(nil)
	s.SetError(nil)
	return s.Exitcode, err
}
