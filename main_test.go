// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package main

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
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
			return args.Server == "" && !args.UseTrustedConnection && args.UserName == ""
		}},
		{[]string{"-c", "MYGO", "-C", "-E", "-i", "file1", "-o", "outfile", "-i", "file2"}, func(args SqlCmdArguments) bool {
			return args.BatchTerminator == "MYGO" && args.TrustServerCertificate && len(args.InputFile) == 2 && strings.HasSuffix(args.OutputFile, "outfile")
		}},
		{[]string{"-U", "someuser", "-d", "somedatabase", "-S", "someserver"}, func(args SqlCmdArguments) bool {
			return args.BatchTerminator == "GO" && !args.TrustServerCertificate && args.UserName == "someuser" && args.DatabaseName == "somedatabase" && args.Server == "someserver"
		}},
		// native sqlcmd allows both -q and -Q but only runs the -Q query and exits. We could make them mutually exclusive if desired.
		{[]string{"-q", "select 1", "-Q", "select 2"}, func(args SqlCmdArguments) bool {
			return args.Server == "" && args.InitialQuery == "select 1" && args.Query == "select 2"
		}},
		{[]string{"-S", "someserver/someinstance"}, func(args SqlCmdArguments) bool {
			return args.Server == "someserver/someinstance"
		}},
		{[]string{"-S", "tcp:someserver,10245"}, func(args SqlCmdArguments) bool {
			return args.Server == "tcp:someserver,10245"
		}},
		{[]string{"-X"}, func(args SqlCmdArguments) bool {
			return args.DisableCmdAndWarn
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
