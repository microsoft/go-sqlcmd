// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"errors"
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/root"
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"os"
	"runtime"
	"strings"
	"testing"
)

// Set to true to run unit tests without a network connection
var offlineMode = false
var useCached = ""
var encryptPassword = ""

type test struct {
	name string
	args struct{ args []string }
}

func init() {
	if runtime.GOOS == "windows" {
		encryptPassword = " --encrypt-password"
	}
}

func TestCommandLineHelp(t *testing.T) {
	setup(t.Name())
	tests := []test{
		{"default", split("--help")},
	}
	run(t, tests)
}

func TestNegCommandLines(t *testing.T) {
	setup(t.Name())
	tests := []test{
		{"neg-config-use-context-double-name",
			split("config use-context badbad --name andbad")},
		{"neg-config-use-context-bad-name",
			split("config use-context badbad")},
		{"neg-config-get-contexts-bad-context",
			split("config get-contexts badbad")},
		{"neg-config-get-endpoints-bad-endpoint",
			split("config get-endpoints badbad")},
		{"neg-install-no-eula",
			split("install mssql")},
	}
	run(t, tests)
}

func TestConfigContexts(t *testing.T) {
	t.Skip()
	setup(t.Name())
	tests := []test{
		{"neg-config-add-context-no-endpoint",
			split("config add-context")},
		{"config-add-endpoint",
			split("config add-endpoint --address localhost --port 1433")},
		{"config-add-endpoint",
			split("config add-endpoint --address localhost --port 1433")},
		{"neg-config-add-context-bad-user",
			split("config add-context --endpoint endpoint --user badbad")},
		{"config-get-endpoints",
			split("config get-endpoints endpoint")},
		{"config-get-endpoints",
			split("config get-endpoints")},
		{"config-get-endpoints",
			split("config get-endpoints --detailed")},
		{"config-add-context",
			split("config add-context --endpoint endpoint")},
		{"uninstall-but-context-has-no-container",
			split("uninstall --force --yes")},
		{"config-add-endpoint",
			split("config add-endpoint")},
		{"config-add-context",
			split("config add-context --endpoint endpoint")},
		{"config-use-context",
			split("config use-context context")},
		{"config-get-contexts",
			split("config get-contexts context")},
		{"config-get-contexts",
			split("config get-contexts")},
		{"config-get-contexts",
			split("config get-contexts --detailed")},
		{"config-delete-context",
			split("config delete-context context --cascade")},
		{"neg-config-delete-context",
			split("config delete-context")},
		{"neg-config-delete-context",
			split("config delete-context badbad-name")},

		{"cleanup",
			split("config delete-endpoint endpoint2")},
		{"cleanup",
			split("config delete-endpoint endpoint3")},
		{"cleanup",
			split("config delete-context context2")},
	}

	run(t, tests)
}

func TestConfigUsers(t *testing.T) {
	setup(t.Name())
	tests := []test{
		{"neg-config-get-users-bad-user",
			split("config get-users badbad")},
		{"config-add-user",
			split("config add-user --username foobar")},
		{"config-add-user",
			split("config add-user --username foobar")},
		{"config-get-users",
			split("config get-users user")},
		{"config-get-users",
			split("config get-users")},
		{"config-get-users",
			split("config get-users --detailed")},
		{"neg-config-add-user-no-username",
			split("config add-user")},
		{"neg-config-add-user-no-password",
			split("config add-user --username foobar")},

		// Cleanup
		{"cleanup",
			split("config delete-user user")},
		{"cleanup",
			split("config delete-user user2")},
	}

	run(t, tests)
}

func TestLocalContext(t *testing.T) {
	setup(t.Name())

	tests := []test{
		{"neg-config-delete-endpoint-no-name",
			split("config delete-endpoint")},
		{"config-add-endpoint",
			split("config add-endpoint --address localhost --port 1433")},
		{"config-add-user",
			split("config add-user --username foobar")},
		{"config-add-context",
			split("config add-context --user user --endpoint endpoint --name my-context")},
		{"config-delete-context-cascade",
			split("config delete-context my-context --cascade")},
		{"config-view",
			split("config view")},
		{"config-view",
			split("config view --raw")},

		{"neg-config-add-user-bad-auth-type",
			split("config add-user --username foobar --auth-type badbad")},
	}

	if len(encryptPassword) > 2 { // are we on a platform that supports encryption
		tests = append(tests, test{"neg-config-add-user-bad-use-encrypted",
			split(fmt.Sprintf("config add-user --username foobar --auth-type other%v", encryptPassword))})
	}

	run(t, tests)
}

func TestGetTags(t *testing.T) {
	setup(t.Name())
	tests := []test{
		{"get-tags",
			split("install mssql get-tags")},
	}

	run(t, tests)
}

func TestMssqlInstall(t *testing.T) {
	setup(t.Name())
	tests := []test{
		{"install",
			split(fmt.Sprintf("install mssql%v --user-database my-database --accept-eula%v", useCached, encryptPassword))},
		{"config-current-context",
			split("config current-context")},
		{"config-connection-strings",
			split("config connection-strings")},
		{"query",
			split("query GO")},
		{"query",
			split("query")},
		{"neg-query-two-queries",
			split("query bad --query bad")},

		/* How to get code coverage for user input
		{"neg-uninstall-no-yes",
			split("uninstall")},*/
		{"uninstall",
			split("uninstall --yes --force")},
	}

	run(t, tests)
}

func runTests(t *testing.T, tt struct {
	name string
	args struct{ args []string }
}) {
	parser := cmdparser.New[*Root](root.SubCommands()...)
	parser.ArgsForUnitTesting(tt.args.args)

	t.Logf("Running: %v", tt.args.args)

	if tt.name == "neg-config-add-user-no-password" {
		os.Setenv("SQLCMD_PASSWORD", "")
	} else {
		os.Setenv("SQLCMD_PASSWORD", "badpass")
	}

	// If test name starts with 'neg-' expect a Panic
	if strings.HasPrefix(tt.name, "neg-") {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		parser.Execute()
	}
	parser.Execute()
}

func Test_displayHints(t *testing.T) {
	root := NewRoot()
	root.displayHints([]string{"Test Hint"})
}

func TestIsValidRootCommand(t *testing.T) {
	root := NewRoot()
	root.IsValidSubCommand("install")
	root.IsValidSubCommand("create")
	root.IsValidSubCommand("nope")
}

func TestRunCommand(t *testing.T) {
	t.Skip()
	root := NewRoot()
	//cmd.loggingLevel = 4
	root.Execute()
}

func Test_checkErr(t *testing.T) {
	t.Skip()
	root := NewRoot()
	//cmd.loggingLevel = 3

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	root.checkErr(errors.New("Expected error"))
}

func run(t *testing.T, tests []test) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runTests(t, tt) })
	}

	verifyConfigIsEmpty(t)
}

func verifyConfigIsEmpty(t *testing.T) {
	if !config.IsEmpty() {
		c := config.GetRedactedConfig(true)
		t.Errorf("Config is not empty. Content of config file:\n%v\nConfig file used:%s",
			c,
			config.GetConfigFileUsed())
		t.Fail()
	}
}

func setup(testName string) {
	useCached = " --cached"
	if !offlineMode {
		useCached = ""
	}

	options := internal.InitializeOptions{
		ErrorHandler: func(err error) {
			if err != nil {
				panic(err)
			}
		},
		HintHandler:  func(i []string) {},
		OutputType:   "yaml",
		LoggingLevel: 4,
	}
	internal.Initialize(options)
	config.SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd",
		"sqlconfig-"+testName,
	))
	config.Clean()
}

type args struct {
	args []string
}

func split(cmd string) args {
	return args{strings.Split(cmd, " ")}
}
