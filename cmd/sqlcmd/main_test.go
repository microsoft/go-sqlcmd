// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package main

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const oneRowAffected = "(1 row affected)"

func newKong(t *testing.T, cli interface{}, options ...kong.Option) *kong.Kong {
	t.Helper()
	options = append([]kong.Option{
		kong.Name("test"),
		kong.Exit(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
	}, options...)
	parser, err := kong.New(cli, options...)
	require.NoError(t, err)
	return parser
}

func TestValidCommandLineToArgsConversion(t *testing.T) {
	type cmdLineTest struct {
		commandLine []string
		check       func(SQLCmdArguments) bool
	}

	// These tests only cover compatibility with the native sqlcmd, which only supports the short flags
	// The long flag names are up for debate.
	commands := []cmdLineTest{
		{[]string{}, func(args SQLCmdArguments) bool {
			return args.Server == "" && !args.UseTrustedConnection && args.UserName == ""
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
		{[]string{"-X", "-x"}, func(args SQLCmdArguments) bool {
			return args.DisableCmdAndWarn && args.DisableVariableSubstitution
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
			return args.ErrorsToStderr == 1
		}},
	}

	for _, test := range commands {
		arguments := &SQLCmdArguments{}
		parser := newKong(t, arguments)
		_, err := parser.Parse(test.commandLine)
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
		{[]string{"-E", "-U", "someuser"}, "--use-trusted-connection and --user-name can't be used together"},
		// the test prefix is a kong artifact https://github.com/alecthomas/kong/issues/221
		{[]string{"-a", "100"}, "test: '-a 100': Packet size has to be a number between 512 and 32767."},
		{[]string{"-F", "what"}, "--format must be one of \"horiz\",\"horizontal\",\"vert\",\"vertical\" but got \"what\""},
		{[]string{"-r", "5"}, `--errors-to-stderr must be one of "-1","0","1" but got '\x05'`},
	}

	for _, test := range commands {
		arguments := &SQLCmdArguments{}
		parser := newKong(t, arguments)
		_, err := parser.Parse(test.commandLine)
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
	vars := sqlcmd.InitializeVariables(!args.DisableCmdAndWarn)
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
	vars := sqlcmd.InitializeVariables(!args.DisableCmdAndWarn)
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
	vars := sqlcmd.InitializeVariables(!args.DisableCmdAndWarn)
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

// Assuming public Azure, use AAD when SQLCMDUSER environment variable is not set
func canTestAzureAuth() bool {
	server := os.Getenv(sqlcmd.SQLCMDSERVER)
	userName := os.Getenv(sqlcmd.SQLCMDUSER)
	return strings.Contains(server, ".database.windows.net") && userName == ""
}
