package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		check       func(SqlCmdArguments) bool
	}

	// These tests only cover compatibility with the native sqlcmd, which only supports the short flags
	// The long flag names are up for debate.
	commands := []cmdLineTest{
		{[]string{}, func(args SqlCmdArguments) bool {
			return args.Server == "." && !args.UseTrustedConnection && args.UserName == ""
		}},
		{[]string{"-c", "MYGO", "-C", "-E", "-i", "file1", "-o", "outfile", "-i", "file2"}, func(args SqlCmdArguments) bool {
			return args.BatchTerminator == "MYGO" && args.TrustServerCertificate && len(args.InputFile) == 2 && strings.HasSuffix(args.OutputFile, "outfile")
		}},
		{[]string{"-U", "someuser", "-d", "somedatabase", "-P", "somestring", "-S", "someserver"}, func(args SqlCmdArguments) bool {
			return args.BatchTerminator == "GO" && !args.TrustServerCertificate && args.UserName == "someuser" && args.DatabaseName == "somedatabase" && args.Password == "somestring" && args.Server == "someserver"
		}},
		// native sqlcmd allows both -q and -Q but only runs the -Q query and exits. We could make them mutually exclusive if desired.
		{[]string{"-q", "select 1", "-Q", "select 2"}, func(args SqlCmdArguments) bool {
			return args.Server == "." && args.InitialQuery == "select 1" && args.Query == "select 2"
		}},
		{[]string{"-S", "someserver/someinstance"}, func(args SqlCmdArguments) bool {
			return args.Server == "someserver/someinstance"
		}},
		{[]string{"-S", "tcp:someserver,10245"}, func(args SqlCmdArguments) bool {
			return args.Server == "tcp:someserver,10245"
		}},
	}

	for _, test := range commands {
		arguments := &SqlCmdArguments{}
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

func TestEnvironmentVariablesAsCommandLineArguments(t *testing.T) {
	os.Setenv("SQLCMDSERVER", "someserver")
	defer os.Unsetenv("SQLCMDSERVER")
	os.Setenv("SQLCMDUSER", "someuser")
	defer os.Unsetenv("SQLCMDUSER")
	os.Setenv("SQLCMDPASSWORD", "1")
	defer os.Unsetenv("SQLCMDPASSWORD")
	arguments := &SqlCmdArguments{}
	parser := newKong(t, arguments)
	_, err := parser.Parse([]string{})
	if assert.NoError(t, err, "Unable to parse empty commandline with environment variables") {
		assert.Equal(t, "someserver", arguments.Server)
		assert.Equal(t, "someuser", arguments.UserName)
		assert.Equal(t, "1", arguments.Password)
	}
}
func TestInvalidCommandLine(t *testing.T) {
	type cmdLineTest struct {
		commandLine  []string
		errorMessage string
	}

	commands := []cmdLineTest{
		{[]string{"-E", "-U", "someuser"}, "--use-trusted-connection and --user-name can't be used together"},
	}

	for _, test := range commands {
		arguments := &SqlCmdArguments{}
		parser := newKong(t, arguments)
		_, err := parser.Parse(test.commandLine)
		assert.EqualError(t, err, test.errorMessage, "Command line:%v", test.commandLine)
	}
}

func TestConnectionStringFromSqlCmdArguments(t *testing.T) {
	type connectionStringTest struct {
		arguments        SqlCmdArguments
		connectionString string
		check            func(*SqlCmdArguments) bool
	}

	pwd := uuid.New().String()

	commands := []connectionStringTest{

		{SqlCmdArguments{Server: "."},
			"sqlserver://.", nil},

		// No username and no explicit setting of UseTrustedConnection will set UseTrustedConnection to true
		{SqlCmdArguments{Server: ".", DatabaseName: "somedatabase", TrustServerCertificate: true},
			"sqlserver://.?database=somedatabase&trustservercertificate=true",
			func(args *SqlCmdArguments) bool {
				return args.UseTrustedConnection
			}},

		{SqlCmdArguments{Server: "someserver/instance", UserName: "someuser", Password: pwd, TrustServerCertificate: true},
			fmt.Sprintf("sqlserver://someuser:%s@someserver/instance?trustservercertificate=true", pwd),
			func(args *SqlCmdArguments) bool {
				return args.Server == "someserver" && args.Port == 0 && args.Instance == "instance"
			}},

		{SqlCmdArguments{Server: "tcp:someserver,1045", UserName: "someuser", Password: pwd, TrustServerCertificate: true},
			fmt.Sprintf("sqlserver://someuser:%s@someserver:1045?trustservercertificate=true", pwd),
			func(args *SqlCmdArguments) bool {
				return args.Server == "someserver" && args.Port == 1045 && args.Instance == ""
			}},
	}

	for _, test := range commands {
		arguments := &test.arguments
		originalArguments := test.arguments
		connectionString, err := connectionString(arguments)
		assert.NoError(t, err)
		assert.Equal(t, test.connectionString, connectionString, "Wrong connection string from: %+v", *arguments)
		if test.check != nil {
			assert.True(t, test.check(arguments), "Unexpected arguments conversion. %+v => %+v", originalArguments, *arguments)
		}
	}
}

func TestConnectionStringErrOnInvalidArguments(t *testing.T) {
	type connectionStringTest struct {
		arguments    SqlCmdArguments
		errorMessage string
	}

	pwd := uuid.New().String()

	commands := []connectionStringTest{
		{SqlCmdArguments{Server: "someserver/instance1/instance2", DatabaseName: "somedatabase", TrustServerCertificate: true, UseTrustedConnection: true},
			"Sqlcmd: Error: server must be of the form [tcp]:server[[/instance]|[,port]]"},

		{SqlCmdArguments{Server: "someserver,notaport", UserName: "someuser", Password: pwd, TrustServerCertificate: true},
			"Sqlcmd: Error: server must be of the form [tcp]:server[[/instance]|[,port]]"},

		{SqlCmdArguments{Server: "tcp:someserver,1045,200", UserName: "someuser", Password: pwd, TrustServerCertificate: true},
			"Sqlcmd: Error: server must be of the form [tcp]:server[[/instance]|[,port]]"},
	}
	for _, test := range commands {
		arguments := &test.arguments
		_, err := connectionString(arguments)
		assert.EqualError(t, err, test.errorMessage, "Wrong error from %+v", test.arguments)
	}
}
