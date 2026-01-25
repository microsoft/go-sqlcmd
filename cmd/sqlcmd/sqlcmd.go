// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
//

package sqlcmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/microsoft/go-mssqldb/msdsn"
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
	// Disable syscommands with a warning or error
	DisableCmd *int
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
	HostNameInCertificate       string
	ServerCertificate           string
	DriverLoggingLevel          int
	ExitOnError                 bool
	ErrorSeverityLevel          uint8
	ErrorLevel                  int
	Vertical                    bool
	ErrorsToStderr              *int
	Headers                     int
	UnicodeOutputFile           bool
	Version                     bool
	ColumnSeparator             string
	ScreenWidth                 *int
	VariableTypeWidth           *int
	FixedTypeWidth              *int
	TrimSpaces                  bool
	Password                    string
	DedicatedAdminConnection    bool
	ListServers                 string
	RemoveControlCharacters     *int
	EchoInput                   bool
	QueryTimeout                int
	EnableColumnEncryption      bool
	ChangePassword              string
	ChangePasswordAndExit       string
	TraceFile                   string
	// PrintStatistics prints performance statistics after each batch
	// nil = disabled, 0 = human-readable, 1 = colon-separated
	PrintStatistics *int
	// Keep Help at the end of the list
	Help bool
}

func (args *SQLCmdArguments) useEnvVars() bool {
	return args.DisableCmd == nil
}

func (args *SQLCmdArguments) errorOnBlockedCmd() bool {
	return args.DisableCmd != nil && *args.DisableCmd > 0
}

func (args *SQLCmdArguments) warnOnBlockedCmd() bool {
	return args.DisableCmd != nil && *args.DisableCmd <= 0
}

func (args *SQLCmdArguments) runStartupScript() bool {
	return args.DisableCmd == nil
}

func (args *SQLCmdArguments) getControlCharacterBehavior() sqlcmd.ControlCharacterBehavior {
	if args.RemoveControlCharacters == nil {
		return sqlcmd.ControlIgnore
	}
	switch *args.RemoveControlCharacters {
	case 1:
		return sqlcmd.ControlReplace
	case 2:
		return sqlcmd.ControlReplaceConsecutive
	}
	return sqlcmd.ControlRemove
}

const (
	sqlcmdErrorPrefix       = "Sqlcmd: "
	applicationIntent       = "application-intent"
	errorsToStderr          = "errors-to-stderr"
	encryptConnection       = "encrypt-connection"
	screenWidth             = "screen-width"
	fixedTypeWidth          = "fixed-type-width"
	variableTypeWidth       = "variable-type-width"
	disableCmdAndWarn       = "disable-cmd-and-warn"
	listServers             = "list-servers"
	removeControlCharacters = "remove-control-characters"
	printStatistics         = "print-statistics"
)

func encryptConnectionAllowsTLS(value string) bool {
	switch strings.ToLower(value) {
	case "s", "strict", "m", "mandatory", "true", "t", "yes", "1":
		return true
	default:
		return false
	}
}

// Validate arguments for settings not describe
func (a *SQLCmdArguments) Validate(c *cobra.Command) (err error) {
	if a.ListServers != "" {
		c.Flags().Visit(func(f *pflag.Flag) {
			if f.Shorthand != "L" {
				err = localizer.Errorf("The -L parameter can not be used in combination with other parameters.")
			}
		})
	}
	if err == nil {
		switch {
		case len(a.InputFile) > 0 && (len(a.Query) > 0 || len(a.InitialQuery) > 0):
			err = mutuallyExclusiveError("i", `-Q/-q`)
		case a.UseTrustedConnection && (len(a.UserName) > 0 || len(a.Password) > 0):
			err = mutuallyExclusiveError("-E", `-U/-P`)
		case a.UseAad && len(a.AuthenticationMethod) > 0:
			err = mutuallyExclusiveError("-G", "--authentication-method")
		case len(a.HostNameInCertificate) > 0 && len(a.ServerCertificate) > 0:
			err = mutuallyExclusiveError("-F", "-J")
		case a.PacketSize != 0 && (a.PacketSize < 512 || a.PacketSize > 32767):
			err = localizer.Errorf(`'-a %#v': Packet size has to be a number between 512 and 32767.`, a.PacketSize)
		// Ignore 0 even though it's technically an invalid input
		case a.Headers < -1:
			err = localizer.Errorf(`'-h %#v': header value must be either -1 or a value between 1 and 2147483647`, a.Headers)
		case a.ScreenWidth != nil && (*a.ScreenWidth < 9 || *a.ScreenWidth > 65535):
			err = rangeParameterError("-w", fmt.Sprint(*a.ScreenWidth), 8, 65536, false)
		case a.FixedTypeWidth != nil && (*a.FixedTypeWidth < 0 || *a.FixedTypeWidth > 8000):
			err = rangeParameterError("-Y", fmt.Sprint(*a.FixedTypeWidth), 0, 8000, true)
		case a.VariableTypeWidth != nil && (*a.VariableTypeWidth < 0 || *a.VariableTypeWidth > 8000):
			err = rangeParameterError("-y", fmt.Sprint(*a.VariableTypeWidth), 0, 8000, true)
		case a.QueryTimeout < 0 || a.QueryTimeout > 65534:
			err = rangeParameterError("-t", fmt.Sprint(a.QueryTimeout), 0, 65534, true)
		case a.ServerCertificate != "" && !encryptConnectionAllowsTLS(a.EncryptConnection):
			err = localizer.Errorf("The -J parameter requires encryption to be enabled (-N true, -N mandatory, or -N strict).")
		}
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
			// emulate -L returning no servers
			if args.ListServers != "" {
				if args.ListServers != "c" {
					fmt.Println()
					fmt.Println(localizer.Sprintf("Servers:"))
				}
				listLocalServers()
				os.Exit(0)
			}
			if len(argss) > 0 {
				fmt.Printf("%s'%s': Unknown command. Enter '--help' for command help.", sqlcmdErrorPrefix, argss[0])
				os.Exit(1)
			}

			vars := sqlcmd.InitializeVariables(args.useEnvVars())
			setVars(vars, &args)

			if args.Version {
				fmt.Println(localizer.ProductBanner())
				fmt.Println()
				fmt.Printf("Version: %v\n", version)
				fmt.Println()
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
		fmt.Println(localizer.ProductBanner())
		fmt.Println()
		fmt.Println(localizer.Sprintf("Version: %v\n", version))
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
	rootCmd.SetArgs(convertOsArgs(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// We need to rewrite the arguments to add -i and -v in front of each space-delimited value to be Cobra-friendly.
// For flags like -r we need to inject the default value if the user omits it
func convertOsArgs(args []string) (cargs []string) {
	flag := ""
	first := true
	for i, a := range args {
		if flag != "" {
			// If the user has a file named "-i" the only way they can pass it on the command line
			// is with triple quotes: sqlcmd -i """-i""" which will convince the flags parser to
			// inject `"-i"` into the string slice. Same for any file with a comma in its name.
			if isFlag(a) {
				flag = ""
			} else if !first {
				cargs = append(cargs, flag)
			}
			first = false
		}
		var defValue string
		if isListFlag(a) {
			flag = a[0:2]
			first = len(a) == 2
		} else {
			defValue = checkDefaultValue(args, i)
		}
		cargs = append(cargs, a)
		if defValue != "" {
			cargs = append(cargs, defValue)
		}
	}
	return
}

// If args[i] is the given flag and args[i+1] is another flag, returns the value to append after the flag
func checkDefaultValue(args []string, i int) (val string) {
	flags := map[rune]string{
		'r': "0",
		'k': "0",
		'L': "|", // | is the sentinel for no value since users are unlikely to use it. It's "reserved" in most shells
		'X': "0",
		'p': "0",
	}
	if isFlag(args[i]) && len(args[i]) == 2 && (len(args) == i+1 || args[i+1][0] == '-') {
		if v, ok := flags[rune(args[i][1])]; ok {
			val = v
			return
		}
	}
	if args[i] == "-N" && (len(args) == i+1 || args[i+1][0] == '-') {
		val = "true"
	}
	return
}

func isFlag(arg string) bool {
	return arg[0] == '-'
}

func isListFlag(arg string) bool {
	return len(arg) > 1 && (arg[0:2] == "-v" || arg[0:2] == "-i")
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
	args.DisableCmd = getOptionalIntArgument(rootCmd, disableCmdAndWarn)
	args.ErrorsToStderr = getOptionalIntArgument(rootCmd, errorsToStderr)
	args.RemoveControlCharacters = getOptionalIntArgument(rootCmd, removeControlCharacters)
	args.PrintStatistics = getOptionalIntArgument(rootCmd, printStatistics)
}

func setFlags(rootCmd *cobra.Command, args *SQLCmdArguments) {
	rootCmd.SetFlagErrorFunc(flagErrorHandler)
	rootCmd.Flags().BoolVarP(&args.Help, "help", "?", false, localizer.Sprintf("-? shows this syntax summary, %s shows modern sqlcmd sub-command help", localizer.HelpFlag))
	rootCmd.Flags().StringVar(&args.TraceFile, "trace-file", "", localizer.Sprintf("Write runtime trace to the specified file. Only for advanced debugging."))
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
	_ = rootCmd.Flags().IntP(disableCmdAndWarn, "X", 0, localizer.Sprintf("%s Disables commands that might compromise system security. Passing 1 tells sqlcmd to exit when disabled commands are run.", "-X[1]"))
	rootCmd.Flags().StringVar(&args.AuthenticationMethod, "authentication-method", "", localizer.Sprintf(
		"Specifies the SQL authentication method to use to connect to Azure SQL Database. One of: %s",
		strings.Join([]string{
			azuread.ActiveDirectoryDefault,
			azuread.ActiveDirectoryIntegrated,
			azuread.ActiveDirectoryPassword,
			azuread.ActiveDirectoryInteractive,
			azuread.ActiveDirectoryManagedIdentity,
			azuread.ActiveDirectoryServicePrincipal,
			azuread.ActiveDirectoryServicePrincipalAccessToken,
			azuread.ActiveDirectoryAzCli,
			azuread.ActiveDirectoryDeviceCode,
			azuread.ActiveDirectoryWorkloadIdentity,
			azuread.ActiveDirectoryClientAssertion,
			azuread.ActiveDirectoryAzurePipelines,
			azuread.ActiveDirectoryEnvironment,
			azuread.ActiveDirectoryAzureDeveloperCli,
			"SqlPassword",
		}, ", "),
	))
	rootCmd.Flags().BoolVarP(&args.UseAad, "use-aad", "G", false, localizer.Sprintf("Tells sqlcmd to use ActiveDirectory authentication. If no user name is provided, authentication method ActiveDirectoryDefault is used. If a password is provided, ActiveDirectoryPassword is used. Otherwise ActiveDirectoryInteractive is used"))
	rootCmd.Flags().BoolVarP(&args.DisableVariableSubstitution, "disable-variable-substitution", "x", false, localizer.Sprintf("Causes sqlcmd to ignore scripting variables. This parameter is useful when a script contains many %s statements that may contain strings that have the same format as regular variables, such as $(variable_name)", localizer.InsertKeyword))
	var variables map[string]string
	rootCmd.Flags().StringToStringVarP(&args.Variables, "variables", "v", variables, localizer.Sprintf("Creates a sqlcmd scripting variable that can be used in a sqlcmd script. Enclose the value in quotation marks if the value contains spaces. You can specify multiple var=values values. If there are errors in any of the values specified, sqlcmd generates an error message and then exits"))
	rootCmd.Flags().IntVarP(&args.PacketSize, "packet-size", "a", 0, localizer.Sprintf("Requests a packet of a different size. This option sets the sqlcmd scripting variable %s. packet_size must be a value between 512 and 32767. The default = 4096. A larger packet size can enhance performance for execution of scripts that have lots of SQL statements between %s commands. You can request a larger packet size. However, if the request is denied, sqlcmd uses the server default for packet size", localizer.PacketSizeVar, localizer.BatchTerminatorGo))
	rootCmd.Flags().IntVarP(&args.LoginTimeout, "login-timeOut", "l", -1, localizer.Sprintf("Specifies the number of seconds before a sqlcmd login to the go-mssqldb driver times out when you try to connect to a server. This option sets the sqlcmd scripting variable %s. The default value is 30. 0 means infinite", localizer.LoginTimeOutVar))
	rootCmd.Flags().StringVarP(&args.WorkstationName, "workstation-name", "H", "", localizer.Sprintf("This option sets the sqlcmd scripting variable %s. The workstation name is listed in the hostname column of the sys.sysprocesses catalog view and can be returned using the stored procedure sp_who. If this option is not specified, the default is the current computer name. This name can be used to identify different sqlcmd sessions", localizer.WorkstationVar))

	rootCmd.Flags().StringVarP(&args.ApplicationIntent, applicationIntent, "K", "default", localizer.Sprintf("Declares the application workload type when connecting to a server. The only currently supported value is ReadOnly. If %s is not specified, the sqlcmd utility will not support connectivity to a secondary replica in an Always On availability group", localizer.ApplicationIntentFlagShort))
	rootCmd.Flags().StringVarP(&args.EncryptConnection, encryptConnection, "N", "default", localizer.Sprintf("This switch is used by the client to request an encrypted connection"))
	rootCmd.Flags().StringVarP(&args.HostNameInCertificate, "host-name-in-certificate", "F", "", localizer.Sprintf("Specifies the host name in the server certificate."))
	rootCmd.Flags().StringVarP(&args.ServerCertificate, "server-certificate", "J", "", localizer.Sprintf("Specifies the path to a server certificate file (PEM, DER, or CER) to match against the server's TLS certificate. Use when encryption is enabled (-N true, -N mandatory, or -N strict) for certificate pinning instead of standard certificate validation."))
	rootCmd.MarkFlagsMutuallyExclusive("host-name-in-certificate", "server-certificate")
	// Can't use NoOptDefVal until this fix: https://github.com/spf13/cobra/issues/866
	//rootCmd.Flags().Lookup(encryptConnection).NoOptDefVal = "true"
	rootCmd.Flags().BoolVarP(&args.Vertical, "vertical", "", false, localizer.Sprintf("Prints the output in vertical format. This option sets the sqlcmd scripting variable %s to '%s'. The default is false", sqlcmd.SQLCMDFORMAT, "vert"))
	_ = rootCmd.Flags().IntP(errorsToStderr, "r", -1, localizer.Sprintf("%s Redirects error messages with severity >= 11 output to stderr. Pass 1 to to redirect all errors including PRINT.", "-r[0 | 1]"))
	rootCmd.Flags().IntVar(&args.DriverLoggingLevel, "driver-logging-level", 0, localizer.Sprintf("Level of mssql driver messages to print"))
	rootCmd.Flags().BoolVarP(&args.ExitOnError, "exit-on-error", "b", false, localizer.Sprintf("Specifies that sqlcmd exits and returns a %s value when an error occurs", localizer.DosErrorLevel))
	rootCmd.Flags().IntVarP(&args.ErrorLevel, "error-level", "m", 0, localizer.Sprintf("Controls which error messages are sent to %s. Messages that have severity level greater than or equal to this level are sent", localizer.StdoutName))

	//Need to decide on short of Header , as "h" is already used in help command in Cobra
	rootCmd.Flags().IntVarP(&args.Headers, "headers", "h", 0, localizer.Sprintf("Specifies the number of rows to print between the column headings. Use -h-1 to specify that headers not be printed"))

	rootCmd.Flags().BoolVarP(&args.UnicodeOutputFile, "unicode-output-file", "u", false, localizer.Sprintf("Specifies that all output files are encoded with little-endian Unicode"))
	rootCmd.Flags().StringVarP(&args.ColumnSeparator, "column-separator", "s", "", localizer.Sprintf("Specifies the column separator character. Sets the %s variable.", localizer.ColSeparatorVar))
	rootCmd.Flags().BoolVarP(&args.TrimSpaces, "trim-spaces", "W", false, localizer.Sprintf("Remove trailing spaces from a column"))
	_ = rootCmd.Flags().BoolP("multi-subnet-failover", "M", false, localizer.Sprintf("Provided for backward compatibility. Sqlcmd always optimizes detection of the active replica of a SQL Failover Cluster"))

	rootCmd.Flags().StringVarP(&args.Password, "password", "P", "", localizer.Sprintf("Password"))

	// Using PersistentFlags() for ErrorSeverityLevel due to data type uint8 , which is not supported in Flags()
	rootCmd.PersistentFlags().Uint8VarP(&args.ErrorSeverityLevel, "error-severity-level", "V", 0, localizer.Sprintf("Controls the severity level that is used to set the %s variable on exit", localizer.ErrorLevel))

	_ = rootCmd.Flags().IntP(screenWidth, "w", 0, localizer.Sprintf("Specifies the screen width for output"))
	_ = rootCmd.Flags().IntP(variableTypeWidth, "y", 256, setScriptVariable("SQLCMDMAXVARTYPEWIDTH"))
	_ = rootCmd.Flags().IntP(fixedTypeWidth, "Y", 0, setScriptVariable("SQLCMDMAXFIXEDTYPEWIDTH"))
	rootCmd.Flags().StringVarP(&args.ListServers, listServers, "L", "", localizer.Sprintf("%s List servers. Pass %s to omit 'Servers:' output.", "-L[c]", "c"))
	rootCmd.Flags().BoolVarP(&args.DedicatedAdminConnection, "dedicated-admin-connection", "A", false, localizer.Sprintf("Dedicated administrator connection"))
	_ = rootCmd.Flags().BoolP("enable-quoted-identifiers", "I", true, localizer.Sprintf("Provided for backward compatibility. Quoted identifiers are always enabled"))
	_ = rootCmd.Flags().BoolP("client-regional-setting", "R", false, localizer.Sprintf("Provided for backward compatibility. Client regional settings are not used"))
	_ = rootCmd.Flags().IntP(removeControlCharacters, "k", 0, localizer.Sprintf("%s Remove control characters from output. Pass 1 to substitute a space per character, 2 for a space per consecutive characters", "-k [1|2]"))
	_ = rootCmd.Flags().IntP(printStatistics, "p", -1, localizer.Sprintf("%s Print performance statistics for every result set. Pass 1 to output in colon-separated format", "-p[1]"))
	rootCmd.Flags().BoolVarP(&args.EchoInput, "echo-input", "e", false, localizer.Sprintf("Echo input"))
	rootCmd.Flags().IntVarP(&args.QueryTimeout, "query-timeout", "t", 0, "Query timeout")
	rootCmd.Flags().BoolVarP(&args.EnableColumnEncryption, "enable-column-encryption", "g", false, localizer.Sprintf("Enable column encryption"))
	rootCmd.Flags().StringVarP(&args.ChangePassword, "change-password", "z", "", localizer.Sprintf("New password"))
	rootCmd.Flags().StringVarP(&args.ChangePasswordAndExit, "change-password-exit", "Z", "", localizer.Sprintf("New password and exit"))
}

func setScriptVariable(v string) string {
	return localizer.Sprintf("Sets the sqlcmd scripting variable %s", v)
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
			case "mandatory", "yes", "1", "t", "true", "disable", "optional", "no", "0", "f", "false", "strict", "m", "s", "o":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-N", v, "m[andatory]", "yes", "1", "t[rue]", "disable", "o[ptional]", "no", "0", "f[alse]", "s[trict]")
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
		case disableCmdAndWarn:
			switch v {
			case "0", "1":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-X", v, "1")
				return pflag.NormalizedName("")
			}
		case listServers:
			switch v {
			case "|", "c":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-L", v, "c")
				return pflag.NormalizedName("")
			}
		case removeControlCharacters:
			switch v {
			case "0", "1", "2":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-k", v, "1", "2")
				return pflag.NormalizedName("")
			}
		case printStatistics:
			switch v {
			case "0", "1":
				return pflag.NormalizedName(name)
			default:
				err = invalidParameterError("-p", v, "0", "1")
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

func mutuallyExclusiveError(flag1 string, flag2 string) error {
	return localizer.Errorf("The %s and the %s options are mutually exclusive.", flag1, flag2)
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
		sqlcmd.SQLCMDSTATTIMEOUT: func(a *SQLCmdArguments) string { return fmt.Sprint(a.QueryTimeout) },
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
		sqlcmd.SQLCMDFORMAT: func(a *SQLCmdArguments) string {
			if a.Vertical {
				return "vert"
			}
			return "horizontal"
		},
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
	} else if args.useEnvVars() {
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
	connect.DisableEnvironmentVariables = !args.useEnvVars()
	connect.DisableVariableSubstitution = args.DisableVariableSubstitution
	connect.ApplicationIntent = args.ApplicationIntent
	connect.LoginTimeoutSeconds = args.LoginTimeout
	switch args.EncryptConnection {
	case "s":
		connect.Encrypt = "strict"
	case "o":
		connect.Encrypt = "optional"
	case "m":
		connect.Encrypt = "mandatory"
	default:
		connect.Encrypt = args.EncryptConnection
	}
	connect.HostNameInCertificate = args.HostNameInCertificate
	connect.ServerCertificate = args.ServerCertificate
	connect.PacketSize = args.PacketSize
	connect.WorkstationName = args.WorkstationName
	connect.LogLevel = args.DriverLoggingLevel
	connect.ExitOnError = args.ExitOnError
	connect.ErrorSeverityLevel = args.ErrorSeverityLevel
	connect.DedicatedAdminConnection = args.DedicatedAdminConnection
	connect.EnableColumnEncryption = args.EnableColumnEncryption
	if len(args.ChangePassword) > 0 {
		connect.ChangePassword = args.ChangePassword
	}
	if len(args.ChangePasswordAndExit) > 0 {
		connect.ChangePassword = args.ChangePasswordAndExit
	}
}

func isConsoleInitializationRequired(connect *sqlcmd.ConnectSettings, args *SQLCmdArguments) (bool, bool) {
	needsConsole := false

	// Check if stdin is from a terminal or a redirection
	isStdinRedirected := false
	file, err := os.Stdin.Stat()
	if err == nil {
		// If stdin is not a character device, it's coming from a pipe or redirect
		if (file.Mode() & os.ModeCharDevice) == 0 {
			isStdinRedirected = true
		}
	}

	// Determine if we're in interactive mode
	iactive := args.InputFile == nil && args.Query == "" && len(args.ChangePasswordAndExit) == 0 && !isStdinRedirected

	// Password input always requires console initialization
	if connect.RequiresPassword() {
		needsConsole = true
	} else if iactive {
		// Interactive mode also requires console
		needsConsole = true
	}

	return needsConsole, iactive
}

func run(vars *sqlcmd.Variables, args *SQLCmdArguments) (int, error) {
	if args.TraceFile != "" {
		traceFile, err := os.Create(args.TraceFile)
		if err != nil {
			return 1, localizer.Errorf("failed to create trace file '%s': %v", args.TraceFile, err)
		}
		defer traceFile.Close()
		err = trace.Start(traceFile)
		if err != nil {
			return 1, localizer.Errorf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	}
	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	var connectConfig sqlcmd.ConnectSettings
	setConnect(&connectConfig, args, vars)
	var line sqlcmd.Console = nil
	needsConsole, isInteractive := isConsoleInitializationRequired(&connectConfig, args)
	if needsConsole {
		line = console.NewConsole("")
		defer line.Close()
	}

	s := sqlcmd.New(line, wd, vars)
	// We want the default behavior on ctrl-c - exit the process
	s.SetupCloseHandler()
	defer s.StopCloseHandler()
	s.UnicodeOutputFile = args.UnicodeOutputFile

	if args.DisableCmd != nil {
		s.Cmd.DisableSysCommands(args.errorOnBlockedCmd())
	}
	s.EchoInput = args.EchoInput
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
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(args.TrimSpaces, args.getControlCharacterBehavior())
	s.PrintStatistics = args.PrintStatistics
	if args.OutputFile != "" {
		err = s.RunCommand(s.Cmd["OUT"], []string{args.OutputFile})
		if err != nil {
			return 1, err
		}
	} else if args.ErrorsToStderr != nil {
		var stderrSeverity uint8 = 11
		if *args.ErrorsToStderr == 1 {
			stderrSeverity = 0
		}

		s.PrintError = func(msg string, severity uint8) bool {
			if severity >= stderrSeverity {
				s.WriteError(os.Stderr, errors.New(msg+sqlcmd.SqlcmdEol))
				return true
			}
			return false
		}
	}

	// connect using no overrides
	err = s.ConnectDb(nil, line == nil)
	if err != nil {
		switch e := err.(type) {
		// 18488 == password must be changed on connection
		case mssql.Error:
			if e.Number == 18488 && line != nil && len(args.Password) == 0 && len(args.ChangePassword) == 0 && len(args.ChangePasswordAndExit) == 0 {
				b, _ := line.ReadPassword(localizer.Sprintf("Enter new password:"))
				s.Connect.ChangePassword = string(b)
				err = s.ConnectDb(nil, true)
			}
		}
		if err != nil {
			s.WriteError(s.GetError(), err)
			return 1, err
		}
	}

	if len(args.ChangePasswordAndExit) > 0 {
		return 0, nil
	}

	script := vars.StartupScriptFile()
	if args.runStartupScript() && len(script) > 0 {
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
		if args.Query != "" {
			once = true
			s.Query = args.Query
		} else if args.InitialQuery != "" {
			s.Query = args.InitialQuery
		}
		iactive := args.InputFile == nil && args.Query == ""
		if iactive || s.Query != "" {
			// If we're not in interactive mode and stdin is redirected,
			// we want to process all input without requiring GO statements
			processAll := !isInteractive
			err = s.Run(once, processAll)
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

func listLocalServers() {
	bmsg := []byte{byte(msdsn.BrowserAllInstances)}
	resp := make([]byte, 16*1024-1)
	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "udp", ":1434")
	// silently ignore failures to connect, same as ODBC
	if err != nil {
		return
	}
	defer conn.Close()
	dl, _ := ctx.Deadline()
	_ = conn.SetDeadline(dl)
	_, err = conn.Write(bmsg)
	if err != nil {
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Println(err)
		}
		return
	}
	read, err := conn.Read(resp)
	if err != nil {
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Println(err)
		}
		return
	}

	data := parseInstances(resp[:read])
	instances := make([]string, 0, len(data))
	for s := range data {
		if s == "MSSQLSERVER" {

			instances = append(instances, "(local)", data[s]["ServerName"])
		} else {
			instances = append(instances, fmt.Sprintf(`%s\%s`, data[s]["ServerName"], s))
		}
	}
	for _, s := range instances {
		fmt.Println("  ", s)
	}
}

func parseInstances(msg []byte) msdsn.BrowserData {
	results := msdsn.BrowserData{}
	if len(msg) > 3 && msg[0] == 5 {
		out_s := string(msg[3:])
		tokens := strings.Split(out_s, ";")
		instdict := map[string]string{}
		got_name := false
		var name string
		for _, token := range tokens {
			if got_name {
				instdict[name] = token
				got_name = false
			} else {
				name = token
				if len(name) == 0 {
					if len(instdict) == 0 {
						break
					}
					results[strings.ToUpper(instdict["InstanceName"])] = instdict
					instdict = map[string]string{}
					continue
				}
				got_name = true
			}
		}
	}
	return results
}
