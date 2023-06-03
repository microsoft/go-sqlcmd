// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
//

package sqlcmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
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
	VariableTypeWidth           *int
	FixedTypeWidth              *int
	TrimSpaces                  bool
	MultiSubnetFailover         bool
	Password                    string
	DedicatedAdminConnection    bool
	// Keep Help at the end of the list
	Help bool
}

const (
	sqlcmdErrorPrefix = "Sqlcmd: "
	applicationIntent = "application-intent"
	errorsToStderr    = "errors-to-stderr"
	format            = "format"
	encryptConnection = "encrypt-connection"
	screenWidth       = "screen-width"
	fixedTypeWidth    = "fixed-type-width"
	variableTypeWidth = "variable-type-width"
)

// Validate arguments for settings not describe
func (a *SQLCmdArguments) Validate(c *cobra.Command) (err error) {
	switch {
	case a.PacketSize != 0 && (a.PacketSize < 512 || a.PacketSize > 32767):
		err = localizer.Errorf(`'-a %d': Packet size has to be a number between 512 and 32767.`, a.PacketSize)
	// Ignore 0 even though it's technically an invalid input
	case a.Headers < -1:
		err = localizer.Errorf(`'-h %d': header value must be either -1 or a value between 1 and 2147483647`, a.Headers)
	case a.ScreenWidth != nil && (*a.ScreenWidth < 9 || *a.ScreenWidth > 65535):
		err = rangeParameterError("-w", fmt.Sprint(*a.ScreenWidth), 8, 65536, false)
	case a.FixedTypeWidth != nil && (*a.FixedTypeWidth < 0 || *a.FixedTypeWidth > 8000):
		err = rangeParameterError("-Y", fmt.Sprint(*a.FixedTypeWidth), 0, 8000, true)
	case a.VariableTypeWidth != nil && (*a.VariableTypeWidth < 0 || *a.VariableTypeWidth > 8000):
		err = rangeParameterError("-y", fmt.Sprint(*a.VariableTypeWidth), 0, 8000, true)
	}
	if err != nil {
		c.PrintErrln(sqlcmdErrorPrefix + err.Error())
		c.SilenceErrors = true
	}
	return
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
// 1. -v: to specify multiple variables. use either "-v var1=v -v var2=v2" or "-v var1=v,var2=v2"
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
	rootCmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, argss []string) error {
			SetScreenWidthFlags(&args, cmd)
			if err := args.Validate(cmd); err != nil {
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
				fmt.Printf("%s'%s': Unknown command. Enter '--help' for command help.", sqlcmdErrorPrefix, argss[0])
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
		fmt.Println(localizer.Sprintf("Version %v\n", version))
		cmd.Flags().SetInterspersed(false)
		fmt.Println(localizer.Sprintf("Flags:"))
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if len(flag.Shorthand) > 0 {
				fmt.Printf("-%s,--%s\n", flag.Shorthand, flag.Name)
			} else {
				fmt.Printf("   --%s\n", flag.Name)
			}
			desc := formatDescription(flag.Usage, 60, 3)
			fmt.Printf("   %s\n", desc)
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

// returns -1 if the parameter has a non-integer value
func getOptionalIntArgument(cmd *cobra.Command, name string) (i *int) {
	i = nil
	val := cmd.Flags().Lookup(name)
	if val != nil && val.Changed {
		i = new(int)
		value := val.Value.String()
		v, e := strconv.Atoi(value)
		if e != nil {
			*i = -1
			return
		}
		*i = v
	}
	return
}

func SetScreenWidthFlags(args *SQLCmdArguments, rootCmd *cobra.Command) {
	args.ScreenWidth = getOptionalIntArgument(rootCmd, screenWidth)
	args.FixedTypeWidth = getOptionalIntArgument(rootCmd, fixedTypeWidth)
	args.VariableTypeWidth = getOptionalIntArgument(rootCmd, variableTypeWidth)
}

func setFlags(rootCmd *cobra.Command, args *SQLCmdArguments) {
	rootCmd.SetFlagErrorFunc(flagErrorHandler)
	rootCmd.Flags().BoolVarP(&args.Help, "help", "?", false, localizer.Sprintf("-? shows this syntax summary, %s shows modern sqlcmd sub-command help", localizer.HelpFlag))
	var inputfiles []string
	rootCmd.Flags().StringSliceVarP(&args.InputFile, "input-file", "i", inputfiles, localizer.Sprintf("Identifies one or more files that contain batches of SQL statements. If one or more files do not exist, sqlcmd will exit. Mutually exclusive with %s/%s", localizer.QueryAndExitFlag, localizer.QueryFlag))
	rootCmd.Flags().StringVarP(&args.OutputFile, "output-file", "o", "", localizer.Sprintf("Identifies the file that receives output from sqlcmd"))
	rootCmd.Flags().BoolVarP(&args.Version, "version", "", false, localizer.Sprintf("Print version information and exit"))
	rootCmd.Flags().BoolVarP(&args.TrustServerCertificate, "trust-server-certificate", "C", false, localizer.Sprintf("Implicitly trust the server certificate without validation"))
	rootCmd.Flags().StringVarP(&args.DatabaseName, "database-name", "d", "", localizer.Sprintf("This option sets the sqlcmd scripting variable %s. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits", localizer.DbNameVar))
	rootCmd.Flags().BoolVarP(&args.UseTrustedConnection, "use-trusted-connection", "E", false, localizer.Sprintf("Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any environment variables that define user name and password"))
	rootCmd.Flags().StringVarP(&args.BatchTerminator, "batch-terminator", "c", "GO", localizer.Sprintf("Specifies the batch terminator. The default value is %s", localizer.BatchTerminatorGo))
	rootCmd.Flags().StringVarP(&args.UserName, "user-name", "U", "", localizer.Sprintf("The login name or contained database user name.  For contained database users, you must provide the database name option"))
	rootCmd.Flags().StringVarP(&args.InitialQuery, "initial-query", "q", "", localizer.Sprintf("Executes a query when sqlcmd starts, but does not exit sqlcmd when the query has finished running. Multiple-semicolon-delimited queries can be executed"))
	rootCmd.Flags().StringVarP(&args.Query, "query", "Q", "", localizer.Sprintf("Executes a query when sqlcmd starts and then immediately exits sqlcmd. Multiple-semicolon-delimited queries can be executed"))
	rootCmd.Flags().StringVarP(&args.Server, "server", "S", "", localizer.Sprintf("%s Specifies the instance of SQL Server to which to connect. It sets the sqlcmd scripting variable %s.", localizer.ConnStrPattern, localizer.ServerEnvVar))
	rootCmd.Flags().BoolVarP(&args.DisableCmdAndWarn, "disable-cmd-and-warn", "X", false, localizer.Sprintf("Disables commands that might compromise system security. Sqlcmd issues a warning and continues"))
	rootCmd.Flags().StringVar(&args.AuthenticationMethod, "authentication-method", "", localizer.Sprintf("Specifies the SQL authentication method to use to connect to Azure SQL Database. One of: ActiveDirectoryDefault, ActiveDirectoryIntegrated, ActiveDirectoryPassword, ActiveDirectoryInteractive, ActiveDirectoryManagedIdentity, ActiveDirectoryServicePrincipal, SqlPassword"))
	rootCmd.Flags().BoolVarP(&args.UseAad, "use-aad", "G", false, localizer.Sprintf("Tells sqlcmd to use ActiveDirectory authentication. If no user name is provided, authentication method ActiveDirectoryDefault is used. If a password is provided, ActiveDirectoryPassword is used. Otherwise ActiveDirectoryInteractive is used"))
	rootCmd.Flags().BoolVarP(&args.DisableVariableSubstitution, "disable-variable-substitution", "x", false, localizer.Sprintf("Causes sqlcmd to ignore scripting variables. This parameter is useful when a script contains many %s statements that may contain strings that have the same format as regular variables, such as $(variable_name)", localizer.InsertKeyword))
	var variables map[string]string
	rootCmd.Flags().StringToStringVarP(&args.Variables, "variables", "v", variables, localizer.Sprintf("Creates a sqlcmd scripting variable that can be used in a sqlcmd script. Enclose the value in quotation marks if the value contains spaces. You can specify multiple var=values values. If there are errors in any of the values specified, sqlcmd generates an error message and then exits"))
	rootCmd.Flags().IntVarP(&args.PacketSize, "packet-size", "a", 0, localizer.Sprintf("Requests a packet of a different size. This option sets the sqlcmd scripting variable %s. packet_size must be a value between 512 and 32767. The default = 4096. A larger packet size can enhance performance for execution of scripts that have lots of SQL statements between %s commands. You can request a larger packet size. However, if the request is denied, sqlcmd uses the server default for packet size", localizer.PacketSizeVar, localizer.BatchTerminatorGo))
	rootCmd.Flags().IntVarP(&args.LoginTimeout, "login-timeOut", "l", -1, localizer.Sprintf("Specifies the number of seconds before a sqlcmd login to the go-mssqldb driver times out when you try to connect to a server. This option sets the sqlcmd scripting variable %s. The default value is 30. 0 means infinite", localizer.LoginTimeOutVar))
	rootCmd.Flags().StringVarP(&args.WorkstationName, "workstation-name", "H", "", localizer.Sprintf("This option sets the sqlcmd scripting variable %s. The workstation name is listed in the hostname column of the sys.sysprocesses catalog view and can be returned using the stored procedure sp_who. If this option is not specified, the default is the current computer name. This name can be used to identify different sqlcmd sessions", localizer.WorkstationVar))

	rootCmd.Flags().StringVarP(&args.ApplicationIntent, applicationIntent, "K", "default", localizer.Sprintf("Declares the application workload type when connecting to a server. The only currently supported value is ReadOnly. If %s is not specified, the sqlcmd utility will not support connectivity to a secondary replica in an Always On availability group", localizer.ApplicationIntentFlagShort))
	rootCmd.Flags().StringVarP(&args.EncryptConnection, encryptConnection, "N", "default", localizer.Sprintf("This switch is used by the client to request an encrypted connection"))
	// Can't use NoOptDefVal until this fix: https://github.com/spf13/cobra/issues/866
	//rootCmd.Flags().Lookup(encryptConnection).NoOptDefVal = "true"
	rootCmd.Flags().StringVarP(&args.Format, format, "F", "horiz", localizer.Sprintf("Specifies the formatting for results"))
	rootCmd.Flags().IntVarP(&args.ErrorsToStderr, errorsToStderr, "r", -1, localizer.Sprintf("Controls which error messages are sent to stdout. Messages that have severity level greater than or equal to this level are sent"))
	//rootCmd.Flags().Lookup(errorsToStderr).NoOptDefVal = "0"
	rootCmd.Flags().IntVar(&args.DriverLoggingLevel, "driver-logging-level", 0, localizer.Sprintf("Level of mssql driver messages to print"))
	rootCmd.Flags().BoolVarP(&args.ExitOnError, "exit-on-error", "b", false, localizer.Sprintf("Specifies that sqlcmd exits and returns a %s value when an error occurs", localizer.DosErrorLevel))
	rootCmd.Flags().IntVarP(&args.ErrorLevel, "error-level", "m", 0, localizer.Sprintf("Controls which error messages are sent to %s. Messages that have severity level greater than or equal to this level are sent", localizer.StdoutName))

	//Need to decide on short of Header , as "h" is already used in help command in Cobra
	rootCmd.Flags().IntVarP(&args.Headers, "headers", "h", 0, localizer.Sprintf("Specifies the number of rows to print between the column headings. Use -h-1 to specify that headers not be printed"))

	rootCmd.Flags().BoolVarP(&args.UnicodeOutputFile, "unicode-output-file", "u", false, localizer.Sprintf("Specifies that all output files are encoded with little-endian Unicode"))
	rootCmd.Flags().StringVarP(&args.ColumnSeparator, "column-separator", "s", "", localizer.Sprintf("Specifies the column separator character. Sets the %s variable.", localizer.ColSeparatorVar))
	rootCmd.Flags().BoolVarP(&args.TrimSpaces, "trim-spaces", "W", false, localizer.Sprintf("Remove trailing spaces from a column"))
	rootCmd.Flags().BoolVarP(&args.MultiSubnetFailover, "multi-subnet-failover", "M", false, localizer.Sprintf("Provided for backward compatibility. Sqlcmd always optimizes detection of the active replica of a SQL Failover Cluster"))

	rootCmd.Flags().StringVarP(&args.Password, "password", "P", "", localizer.Sprintf("Password"))

	// Using PersistentFlags() for ErrorSeverityLevel due to data type uint8 , which is not supported in Flags()
	rootCmd.PersistentFlags().Uint8VarP(&args.ErrorSeverityLevel, "error-severity-level", "V", 0, localizer.Sprintf("Controls the severity level that is used to set the %s variable on exit", localizer.ErrorLevel))

	_ = rootCmd.Flags().IntP(screenWidth, "w", 0, localizer.Sprintf("Specifies the screen width for output"))
	_ = rootCmd.Flags().IntP(variableTypeWidth, "y", 256, localizer.Sprintf("Sets the sqlcmd scripting variable %s", "SQLCMDMAXVARTYPEWIDTH"))
	_ = rootCmd.Flags().IntP(fixedTypeWidth, "Y", 0, localizer.Sprintf("Sets the sqlcmd scripting variable %s", "SQLCMDMAXFIXEDTYPEWIDTH"))
	rootCmd.Flags().BoolVarP(&args.DedicatedAdminConnection, "dedicated-admin-connection", "A", false, localizer.Sprintf("Dedicated administrator connection"))
}

func normalizeFlags(cmd *cobra.Command) error {
	//Adding a validator for checking the enum flags
	var err error
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		v := getFlagValueByName(f, name)
		if v == "" {
			return pflag.NormalizedName("")
		}
		switch name {
		case applicationIntent:
			value := strings.ToLower(v)
			switch value {
			case "readonly":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError(localizer.ApplicationIntentFlagShort, v, "readonly")
				return pflag.NormalizedName("")
			}
		case encryptConnection:
			value := strings.ToLower(v)
			switch value {
			case "false", "true", "disable":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-N", v, "false", "true", "disable")
				return pflag.NormalizedName("")
			}
		case format:
			value := strings.ToLower(v)
			switch value {
			case "horiz", "horizontal", "vert", "vertical":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-F", v, "horiz", "horizontal", "vert", "vertical")
				return pflag.NormalizedName("")
			}
		case errorsToStderr:
			switch v {
			case "0", "1":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-r", v, "0", "1")
				return pflag.NormalizedName("")
			}
		}
		return pflag.NormalizedName(name)
	})
	if err != nil {
		cmd.PrintErrln(sqlcmdErrorPrefix, err.Error())
		cmd.SilenceErrors = true
	}
	return err
}

var invalidArgRegexp = regexp.MustCompile(`invalid argument \"(.*)\" for \"(-.), (--.*)\"`)
var missingArgRegexp = regexp.MustCompile(`flag needs an argument: '.' in (-.)`)
var unknownArgRegexp = regexp.MustCompile(`unknown shorthand flag: '(.)' in -.`)

func rangeParameterError(flag string, value string, min int, max int, inclusive bool) error {
	if inclusive {
		return localizer.Errorf(`'%s %s': value must be greater than or equal to %#v and less than or equal to %#v.`, flag, value, min, max)
	}
	return localizer.Errorf(`'%s %s': value must be greater than %#v and less than %#v.`, flag, value, min, max)
}

func invalidParameterError(flag string, value string, allowedValues ...string) error {
	if len(allowedValues) == 1 {
		return localizer.Errorf("'%s %s': Unexpected argument. Argument value has to be %v.", flag, value, allowedValues[0])
	}
	return localizer.Errorf("'%s %s': Unexpected argument. Argument value has to be one of %v.", flag, value, allowedValues)
}

func flagErrorHandler(c *cobra.Command, err error) (e error) {
	c.SilenceUsage = true
	c.SilenceErrors = true
	e = nil
	p := invalidArgRegexp.FindStringSubmatch(err.Error())
	if len(p) == 4 {
		f := p[2]
		v := p[1]
		switch f {
		case "-y", "-Y":
			e = rangeParameterError(f, v, 0, 8000, true)
		case "-w":
			e = rangeParameterError(f, v, 8, 65536, false)
		}
	}
	if e == nil {
		p = missingArgRegexp.FindStringSubmatch(err.Error())
		if len(p) == 2 {
			e = localizer.Errorf(`'%s': Missing argument. Enter '-?' for help.`, p[1])
		}
	}
	if e == nil {
		p = unknownArgRegexp.FindStringSubmatch(err.Error())
		if len(p) == 2 {
			e = localizer.Errorf(`'%s': Unknown Option. Enter '-?' for help.`, p[1])
		}
	}
	if e == nil {
		e = err
	}
	c.PrintErrln(sqlcmdErrorPrefix, e.Error())
	return
}

// Returns the value of the flag if it was provided, empty string if it was not provided
func getFlagValueByName(flagSet *pflag.FlagSet, name string) string {
	var value string
	flagSet.Visit(func(f *pflag.Flag) {
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
		sqlcmd.SQLCMDMAXVARTYPEWIDTH: func(a *SQLCmdArguments) string {
			if a.VariableTypeWidth != nil {
				return fmt.Sprint(*a.VariableTypeWidth)
			}
			return ""
		},
		sqlcmd.SQLCMDMAXFIXEDTYPEWIDTH: func(a *SQLCmdArguments) string {
			if a.FixedTypeWidth != nil {
				return fmt.Sprint(*a.FixedTypeWidth)
			}
			return ""
		},
		sqlcmd.SQLCMDFORMAT: func(a *SQLCmdArguments) string { return a.Format },
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
	if len(args.Password) > 0 {
		connect.Password = args.Password
	} else if !args.DisableCmdAndWarn {
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
	connect.DedicatedAdminConnection = args.DedicatedAdminConnection
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
