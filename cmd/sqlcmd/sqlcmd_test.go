// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const oneRowAffected = "(1 row affected)"

func TestValidCommandLineToArgsConversion(t *testing.T) {
	type cmdLineTest struct {
		commandLine []string
		check       func(SQLCmdArguments) bool
	}

	// These tests only cover compatibility with the native sqlcmd, which only supports the short flags
	// The long flag names are up for debate.
	commands := []cmdLineTest{
		{[]string{}, func(args SQLCmdArguments) bool {
			return args.Server == "" && !args.UseTrustedConnection && args.UserName == "" && args.ScreenWidth == nil && args.ErrorsToStderr == nil && args.EncryptConnection == "default"
		}},
		{[]string{"-v", "a=b", "x=y", "-E"}, func(args SQLCmdArguments) bool {
			return len(args.Variables) == 2 && args.Variables["a"] == "b" && args.Variables["x"] == "y" && args.UseTrustedConnection
		}},
		{[]string{"-c", "MYGO", "-C", "-E", "-i", "file1", "-o", "outfile", "-i", "file2"}, func(args SQLCmdArguments) bool {
			return args.BatchTerminator == "MYGO" && args.TrustServerCertificate && len(args.InputFile) == 2 && strings.HasSuffix(args.OutputFile, "outfile")
		}},
		{[]string{"-U", "someuser", "-d", "somedatabase", "-S", "someserver"}, func(args SQLCmdArguments) bool {
			return args.BatchTerminator == "GO" && !args.TrustServerCertificate && args.UserName == "someuser" && args.DatabaseName == "somedatabase" && args.Server == "someserver"
		}},
		// native sqlcmd allows both -q and -Q but only runs the -Q query and exits. We could make them mutually exclusive if desired.
		{[]string{"-q", "select 1", "-Q", "select 2"}, func(args SQLCmdArguments) bool {
			return args.Server == "" && args.InitialQuery == "select 1" && args.Query == "select 2"
		}},
		{[]string{"-S", "someserver/someinstance"}, func(args SQLCmdArguments) bool {
			return args.Server == "someserver/someinstance"
		}},
		{[]string{"-S", "tcp:someserver,10245"}, func(args SQLCmdArguments) bool {
			return args.Server == "tcp:someserver,10245" && !args.DisableVariableSubstitution
		}},
		{[]string{"-X", "1", "-x"}, func(args SQLCmdArguments) bool {
			return args.errorOnBlockedCmd() && args.DisableVariableSubstitution
		}},
		// Notice no "" around the value with a space in it. It seems quotes get stripped out somewhere before Parse when invoking on a real command line
		{[]string{"-v", "x=y", "-v", `y=a space`}, func(args SQLCmdArguments) bool {
			return args.LoginTimeout == -1 && args.Variables["x"] == "y" && args.Variables["y"] == "a space"
		}},
		{[]string{"-a", "550", "-l", "45", "-H", "mystation", "-K", "ReadOnly", "-N", "true"}, func(args SQLCmdArguments) bool {
			return args.PacketSize == 550 && args.LoginTimeout == 45 && args.WorkstationName == "mystation" && args.ApplicationIntent == "ReadOnly" && args.EncryptConnection == "true"
		}},
		{[]string{"-b", "-m", "15", "-V", "20"}, func(args SQLCmdArguments) bool {
			return args.ExitOnError && args.ErrorLevel == 15 && args.ErrorSeverityLevel == 20
		}},
		{[]string{"-F", "vert"}, func(args SQLCmdArguments) bool {
			return args.Format == "vert"
		}},
		{[]string{"-r", "1"}, func(args SQLCmdArguments) bool {
			return *args.ErrorsToStderr == 1
		}},
		{[]string{"-h", "2", "-?"}, func(args SQLCmdArguments) bool {
			return args.Help && args.Headers == 2
		}},
		{[]string{"-u", "-A"}, func(args SQLCmdArguments) bool {
			return args.UnicodeOutputFile && args.DedicatedAdminConnection
		}},
		{[]string{"--version"}, func(args SQLCmdArguments) bool {
			return args.Version
		}},
		{[]string{"-w", "10"}, func(args SQLCmdArguments) bool {
			return args.ScreenWidth != nil && *args.ScreenWidth == 10 && args.FixedTypeWidth == nil && args.VariableTypeWidth == nil
		}},
		{[]string{"-s", "|", "-w", "10", "-W", "-t", "10"}, func(args SQLCmdArguments) bool {
			return args.TrimSpaces && args.ColumnSeparator == "|" && *args.ScreenWidth == 10 && args.QueryTimeout == 10
		}},
		{[]string{"-y", "100", "-Y", "200", "-P", "placeholder", "-e"}, func(args SQLCmdArguments) bool {
			return *args.FixedTypeWidth == 200 && *args.VariableTypeWidth == 100 && args.Password == "placeholder" && args.EchoInput
		}},
		{[]string{"-E", "-v", "a=b", "x=y", "-i", "a.sql", "b.sql", "-v", "f=g", "-i", "c.sql", "-C", "-v", "ab=cd", "ef=hi"}, func(args SQLCmdArguments) bool {
			return args.UseTrustedConnection && args.Variables["x"] == "y" && len(args.InputFile) == 3 && args.InputFile[0] == "a.sql" && args.TrustServerCertificate
		}},
		{[]string{"-i", `comma,text.sql`}, func(args SQLCmdArguments) bool {
			return args.InputFile[0] == "comma" && args.InputFile[1] == "text.sql"
		}},
		{[]string{"-i", `"comma,text.sql"`}, func(args SQLCmdArguments) bool {
			return args.InputFile[0] == "comma,text.sql"
		}},
		{[]string{"-k", "-X", "-r", "-z", "something"}, func(args SQLCmdArguments) bool {
			return args.warnOnBlockedCmd() && !args.useEnvVars() && args.getControlCharacterBehavior() == sqlcmd.ControlRemove && *args.ErrorsToStderr == 0 && args.ChangePassword == "something"
		}},
		{[]string{"-N"}, func(args SQLCmdArguments) bool {
			return args.EncryptConnection == "true"
		}},
		{[]string{"-N", "m"}, func(args SQLCmdArguments) bool {
			return args.EncryptConnection == "m"
		}},
		{[]string{"-ifoo.sql", "bar.sql", "-V10"}, func(args SQLCmdArguments) bool {
			return args.ErrorSeverityLevel == 10 && args.InputFile[0] == "foo.sql" && args.InputFile[1] == "bar.sql"
		}},
		{[]string{"-N", "s:myserver.domain.com"}, func(args SQLCmdArguments) bool {
			return args.EncryptConnection == "s:myserver.domain.com"
		}},
	}

	for _, test := range commands {
		arguments := &SQLCmdArguments{}
		cmd := &cobra.Command{
			Use:   "testCommand",
			Short: "A brief description of my command",
			Long:  "A long description of my command",
			PreRunE: func(cmd *cobra.Command, argss []string) error {
				SetScreenWidthFlags(arguments, cmd)
				return arguments.Validate(cmd)
			},
			Run: func(cmd *cobra.Command, argss []string) {
				// Command logic goes here
			},
			SilenceErrors: true,
			SilenceUsage:  true,
		}
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		setFlags(cmd, arguments)
		cmd.SetArgs(convertOsArgs(test.commandLine))
		err := cmd.Execute()
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if assert.Nil(t, err, "Unable to parse commandLine:%v\n%s", test.commandLine, msg) {
			assert.True(t, test.check(*arguments), "Unexpected SqlCmdArguments from: %v\n%+v", test.commandLine, *arguments)
		}
	}
}

func TestInvalidCommandLine(t *testing.T) {
	type cmdLineTest struct {
		commandLine  []string
		errorMessage string
	}

	commands := []cmdLineTest{
		{[]string{"-E", "-U", "someuser"}, "The -E and the -U/-P options are mutually exclusive."},
		{[]string{"-L", "-q", `"select 1"`}, "The -L parameter can not be used in combination with other parameters."},
		{[]string{"-i", "foo.sql", "-q", `"select 1"`}, "The i and the -Q/-q options are mutually exclusive."},
		{[]string{"-F", "what"}, "'-F what': Unexpected argument. Argument value has to be one of [horiz horizontal vert vertical]."},
		{[]string{"-r", "5"}, `'-r 5': Unexpected argument. Argument value has to be one of [0 1].`},
		{[]string{"-w", "x"}, "'-w x': value must be greater than 8 and less than 65536."},
		{[]string{"-y", "111111"}, "'-y 111111': value must be greater than or equal to 0 and less than or equal to 8000."},
		{[]string{"-Y", "-2"}, "'-Y -2': value must be greater than or equal to 0 and less than or equal to 8000."},
		{[]string{"-P"}, "'-P': Missing argument. Enter '-?' for help."},
		{[]string{"-;"}, "';': Unknown Option. Enter '-?' for help."},
		{[]string{"-t", "-2"}, "'-t -2': value must be greater than or equal to 0 and less than or equal to 65534."},
		{[]string{"-N", "invalid"}, "'-N invalid': Unexpected argument. Argument value has to be one of [m[andatory] yes 1 t[rue] disable o[ptional] no 0 f[alse] s[trict][:<hostnameincertificate>]]."},
	}

	for _, test := range commands {
		arguments := &SQLCmdArguments{}
		cmd := &cobra.Command{
			Use:   "testCommand",
			Short: "A brief description of my command",
			Long:  "A long description of my command",
			PreRunE: func(cmd *cobra.Command, argss []string) error {
				SetScreenWidthFlags(arguments, cmd)
				if err := arguments.Validate(cmd); err != nil {
					cmd.SilenceUsage = true
					return err
				}
				return normalizeFlags(cmd)
			},
			Run: func(cmd *cobra.Command, argss []string) {
			},
			SilenceUsage: true,
		}
		buf := &memoryBuffer{buf: new(bytes.Buffer)}
		cmd.SetErr(buf)
		setFlags(cmd, arguments)
		cmd.SetArgs(convertOsArgs(test.commandLine))
		err := cmd.Execute()
		if assert.EqualErrorf(t, err, test.errorMessage, "Command line: %s", test.commandLine) {
			errBytes := buf.buf.String()
			assert.Equalf(t, sqlcmdErrorPrefix, string(errBytes)[:len(sqlcmdErrorPrefix)], "Output error should start with '%s' but got '%s' - %s", sqlcmdErrorPrefix, string(errBytes), test.commandLine)
		}
	}
}

func TestValidateFlags(t *testing.T) {
	type cmdLineTest struct {
		commandLine  []string
		errorMessage string
	}

	commands := []cmdLineTest{
		{[]string{"-a", "100"}, "'-a 100': Packet size has to be a number between 512 and 32767."},
		{[]string{"-h-4"}, "'-h -4': header value must be either -1 or a value between 1 and 2147483647"},
		{[]string{"-w", "6"}, "'-w 6': value must be greater than 8 and less than 65536."},
	}

	for _, test := range commands {
		arguments := &SQLCmdArguments{}
		//var screenWidth *int
		cmd := &cobra.Command{
			Use:   "testCommand",
			Short: "A brief description of my command",
			Long:  "A long description of my command",
			PreRunE: func(cmd *cobra.Command, argss []string) error {
				SetScreenWidthFlags(arguments, cmd)
				return arguments.Validate(cmd)
			},
			Run: func(cmd *cobra.Command, argss []string) {
			},
			SilenceErrors: true,
			SilenceUsage:  true,
		}
		cmd.SetErr(new(bytes.Buffer))
		setFlags(cmd, arguments)
		cmd.SetArgs(convertOsArgs(test.commandLine))
		err := cmd.Execute()
		assert.EqualError(t, err, test.errorMessage, "Command line:%v", test.commandLine)
	}
}

// Simulate main() using files
func TestRunInputFiles(t *testing.T) {
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.InputFile = []string{"testdata/select100.sql", "testdata/select100.sql"}
	args.OutputFile = o.Name()
	if canTestAzureAuth() {
		args.UseAad = true
	}
	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	vars.Set(sqlcmd.SQLCMDMAXVARTYPEWIDTH, "0")
	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		assert.Equal(t, "100"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol+"100"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol, string(bytes), "Incorrect output from run")
	}
}

func TestUnicodeOutput(t *testing.T) {
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.InputFile = []string{"testdata/selectutf8.txt"}
	args.OutputFile = o.Name()
	args.UnicodeOutputFile = true
	if canTestAzureAuth() {
		args.UseAad = true
	}
	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		outfile := `testdata/unicodeout_linux.txt`
		if runtime.GOOS == "windows" {
			outfile = `testdata/unicodeout.txt`
		}
		expectedBytes, err := os.ReadFile(outfile)
		if assert.NoErrorf(t, err, "Unable to open %s", outfile) {
			assert.Equalf(t, expectedBytes, bytes, "unicode output bytes should match %s", outfile)
		}
	}
}

func TestUnicodeInput(t *testing.T) {
	testfiles := []string{
		filepath.Join(`testdata`, `selectutf8.txt`),
		filepath.Join(`testdata`, `selectutf8_bom.txt`),
		filepath.Join(`testdata`, `selectunicode_BE.txt`),
		filepath.Join(`testdata`, `selectunicode_LE.txt`),
	}

	for _, test := range testfiles {
		for _, unicodeOutput := range []bool{true, false} {
			var outfile string
			if unicodeOutput {
				outfile = filepath.Join(`testdata`, `unicodeout_linux.txt`)
				if runtime.GOOS == "windows" {
					outfile = filepath.Join(`testdata`, `unicodeout.txt`)
				}
			} else {
				outfile = `testdata/utf8out_linux.txt`
				if runtime.GOOS == "windows" {
					outfile = filepath.Join(`testdata`, `utf8out.txt`)
				}
			}
			o, err := os.CreateTemp("", "sqlcmdmain")
			assert.NoError(t, err, "os.CreateTemp")
			defer os.Remove(o.Name())
			defer o.Close()
			args = newArguments()
			args.InputFile = []string{test}
			args.OutputFile = o.Name()
			args.UnicodeOutputFile = unicodeOutput
			if canTestAzureAuth() {
				args.UseAad = true
			}
			vars := sqlcmd.InitializeVariables(args.useEnvVars())
			setVars(vars, &args)
			exitCode, err := run(vars, &args)
			assert.NoError(t, err, "run")
			assert.Equal(t, 0, exitCode, "exitCode")
			bytes, err := os.ReadFile(o.Name())
			s := string(bytes)
			if assert.NoError(t, err, "os.ReadFile") {
				expectedBytes, err := os.ReadFile(outfile)
				expectedS := string(expectedBytes)
				if assert.NoErrorf(t, err, "Unable to open %s", outfile) {
					assert.Equalf(t, expectedS, s, "input file: <%s> output bytes should match <%s>", test, outfile)
				}
			}
		}
	}
}

func TestQueryAndExit(t *testing.T) {
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.Query = "SELECT '$(VAR1) $(VAR2)'"
	args.OutputFile = o.Name()
	args.Variables = map[string]string{"var2": "val2"}
	if canTestAzureAuth() {
		args.UseAad = true
	}
	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	vars.Set(sqlcmd.SQLCMDMAXVARTYPEWIDTH, "0")
	vars.Set("VAR1", "100")
	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		assert.Equal(t, "100 val2"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol, string(bytes), "Incorrect output from run")
	}
}

func TestInitQueryAndQueryExecutesQuery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic occurred: %v", r)
		}
	}()
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.InitialQuery = "SELECT 1"
	args.Query = "SELECT 2"
	args.OutputFile = o.Name()
	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	vars.Set(sqlcmd.SQLCMDMAXVARTYPEWIDTH, "0")

	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		assert.Equal(t, "2"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol, string(bytes), "Incorrect output from run")
	}
}

// Test to verify fix for issue: https://github.com/microsoft/go-sqlcmd/issues/98
//  1. Verify when -b is passed in (ExitOnError), we don't always get an error (even when input is good)
//     2, Verify when the input is actually bad, we do get an error
func TestExitOnError(t *testing.T) {
	args = newArguments()
	args.InputFile = []string{"testdata/select100.sql"}
	args.ErrorsToStderr = new(int)
	*args.ErrorsToStderr = 0
	args.ExitOnError = true
	if canTestAzureAuth() {
		args.UseAad = true
	}

	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")

	args.InputFile = []string{"testdata/bad.sql"}

	vars = sqlcmd.InitializeVariables(args.useEnvVars())
	setVars(vars, &args)

	exitCode, err = run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 1, exitCode, "exitCode")

	t.Logf("Test Completed") // Needs some output to stdout to count as a test
}

func TestAzureAuth(t *testing.T) {
	if !canTestAzureAuth() {
		t.Skip("Server name is not an Azure DB name")
	}
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.Query = "SELECT 'AZURE'"
	args.OutputFile = o.Name()
	args.UseAad = true
	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	vars.Set(sqlcmd.SQLCMDMAXVARTYPEWIDTH, "0")
	setVars(vars, &args)
	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		assert.Equal(t, "AZURE"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol, string(bytes), "Incorrect output from run")
	}
}

func TestMissingInputFile(t *testing.T) {
	args = newArguments()
	args.InputFile = []string{filepath.Join("testdata", "missingFile.sql")}

	if canTestAzureAuth() {
		args.UseAad = true
	}

	vars := sqlcmd.InitializeVariables(args.useEnvVars())
	setVars(vars, &args)

	exitCode, err := run(vars, &args)
	assert.Error(t, err, "run")
	assert.Contains(t, err.Error(), "Error occurred while opening or operating on file", "Unexpected error: "+err.Error())
	assert.Equal(t, 1, exitCode, "exitCode")
}

func TestConditionsForPasswordPrompt(t *testing.T) {

	type test struct {
		authenticationMethod string
		inputFile            []string
		username             string
		pwd                  string
		expectedResult       bool
	}
	tests := []test{
		// Positive Testcases
		{sqlcmd.SqlPassword, []string{""}, "someuser", "", true},
		{sqlcmd.NotSpecified, []string{"testdata/someFile.sql"}, "someuser", "", true},
		{azuread.ActiveDirectoryPassword, []string{""}, "someuser", "", true},
		{azuread.ActiveDirectoryPassword, []string{"testdata/someFile.sql"}, "someuser", "", true},
		{azuread.ActiveDirectoryServicePrincipal, []string{""}, "someuser", "", true},
		{azuread.ActiveDirectoryServicePrincipal, []string{"testdata/someFile.sql"}, "someuser", "", true},
		{azuread.ActiveDirectoryApplication, []string{""}, "someuser", "", true},
		{azuread.ActiveDirectoryApplication, []string{"testdata/someFile.sql"}, "someuser", "", true},

		//Negative Testcases
		{sqlcmd.NotSpecified, []string{""}, "", "", false},
		{sqlcmd.NotSpecified, []string{"testdata/someFile.sql"}, "", "", false},
		{azuread.ActiveDirectoryDefault, []string{""}, "someuser", "", false},
		{azuread.ActiveDirectoryDefault, []string{"testdata/someFile.sql"}, "someuser", "", false},
		{azuread.ActiveDirectoryInteractive, []string{""}, "someuser", "", false},
		{azuread.ActiveDirectoryInteractive, []string{"testdata/someFile.sql"}, "someuser", "", false},
		{azuread.ActiveDirectoryManagedIdentity, []string{""}, "someuser", "", false},
		{azuread.ActiveDirectoryManagedIdentity, []string{"testdata/someFile.sql"}, "someuser", "", false},
	}

	for _, testcase := range tests {
		t.Log(testcase.authenticationMethod, testcase.inputFile, testcase.username, testcase.pwd, testcase.expectedResult)
		args := newArguments()
		args.DisableCmd = new(int)
		*args.DisableCmd = 1
		args.InputFile = testcase.inputFile
		args.UserName = testcase.username
		vars := sqlcmd.InitializeVariables(args.useEnvVars())
		setVars(vars, &args)
		var connectConfig sqlcmd.ConnectSettings
		setConnect(&connectConfig, &args, vars)
		connectConfig.AuthenticationMethod = testcase.authenticationMethod
		connectConfig.Password = testcase.pwd
		needsConsole, _ := isConsoleInitializationRequired(&connectConfig, &args)
		assert.Equal(t, testcase.expectedResult, needsConsole, "Unexpected test result encountered for console initialization")
		assert.Equal(t, testcase.expectedResult, connectConfig.RequiresPassword() && connectConfig.Password == "", "Unexpected test result encountered for password prompt conditions")
	}
}

func TestStartupScript(t *testing.T) {
	o, err := os.CreateTemp("", "sqlcmdmain")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(o.Name())
	defer o.Close()
	args = newArguments()
	args.OutputFile = o.Name()
	args.Query = "set nocount on"
	if canTestAzureAuth() {
		args.UseAad = true
	}
	vars := sqlcmd.InitializeVariables(true)
	setVars(vars, &args)
	vars.Set(sqlcmd.SQLCMDINI, "testdata/select100.sql")
	vars.Set(sqlcmd.SQLCMDMAXVARTYPEWIDTH, "0")
	exitCode, err := run(vars, &args)
	assert.NoError(t, err, "run")
	assert.Equal(t, 0, exitCode, "exitCode")
	bytes, err := os.ReadFile(o.Name())
	if assert.NoError(t, err, "os.ReadFile") {
		assert.Equal(t, "100"+sqlcmd.SqlcmdEol+sqlcmd.SqlcmdEol+oneRowAffected+sqlcmd.SqlcmdEol, string(bytes), "Incorrect output from run")
	}
}

func TestConvertOsArgs(t *testing.T) {
	type test struct {
		name     string
		in       []string
		expected []string
	}

	tests := []test{
		{
			"Multiple variables/one switch",
			[]string{"-E", "-v", "a=b", "x=y", "f=g", "-C"},
			[]string{"-E", "-v", "a=b", "-v", "x=y", "-v", "f=g", "-C"},
		},
		{
			"Multiple variables and files/multiple switches",
			[]string{"-E", "-v", "a=b", "x=y", "-i", "a.sql", "b.sql", "-v", "f=g", "-i", "c.sql", "-C", "-v", "ab=cd", "ef=hi"},
			[]string{"-E", "-v", "a=b", "-v", "x=y", "-i", "a.sql", "-i", "b.sql", "-v", "f=g", "-i", "c.sql", "-C", "-v", "ab=cd", "-v", "ef=hi"},
		},
		{
			"Flags with optional arguments",
			[]string{"-r", "1", "-X", "-k", "-C"},
			[]string{"-r", "1", "-X", "0", "-k", "0", "-C"},
		},
		{
			"-i followed by flags without spaces",
			[]string{"-i", "a.sql", "-V10", "-C"},
			[]string{"-i", "a.sql", "-V10", "-C"},
		},
		{
			"list flags without spaces",
			[]string{"-ifoo.sql", "bar.sql", "-V10", "-X", "-va=b", "c=d"},
			[]string{"-ifoo.sql", "-i", "bar.sql", "-V10", "-X", "0", "-va=b", "-v", "c=d"},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			actual := convertOsArgs(c.in)
			assert.ElementsMatch(t, c.expected, actual, "Incorrect converted args")
		})
	}
}

func TestEncryptionOptions(t *testing.T) {
	type test struct {
		input                 string
		output                string
		hostnameincertificate string
	}
	tests := []test{
		{
			"s",
			"strict",
			"",
		},
		{
			"m",
			"mandatory",
			"",
		},
		{
			"o",
			"optional",
			"",
		},
		{
			"mandatory",
			"mandatory",
			"",
		},
		{
			"s:myserver.domain.com",
			"strict",
			"myserver.domain.com",
		},
		{
			"strict:myserver.domain.com",
			"strict",
			"myserver.domain.com",
		},
	}
	for _, c := range tests {
		t.Run(c.input, func(t *testing.T) {
			args := newArguments()
			args.EncryptConnection = c.input
			vars := sqlcmd.InitializeVariables(false)
			setVars(vars, &args)
			var connectConfig sqlcmd.ConnectSettings
			setConnect(&connectConfig, &args, vars)
			assert.Equal(t, c.output, connectConfig.Encrypt, "Incorrect connect.Encrypt")
			assert.Equal(t, c.hostnameincertificate, connectConfig.HostNameInCertificate, "Incorrect connect.HostNameInCertificate")
		})
	}
}

// Assuming public Azure, use AAD when SQLCMDUSER environment variable is not set
func canTestAzureAuth() bool {
	server := os.Getenv(sqlcmd.SQLCMDSERVER)
	userName := os.Getenv(sqlcmd.SQLCMDUSER)
	return strings.Contains(server, ".database.windows.net") && userName == ""
}

// memoryBuffer has both Write and Close methods for use as io.WriteCloser
type memoryBuffer struct {
	buf *bytes.Buffer
}

func (b *memoryBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

func (b *memoryBuffer) Close() error {
	return nil
}
