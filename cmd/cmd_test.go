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
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/spf13/viper"
	"os"
	"strings"
	"testing"
)

// Set to true to run unit tests without a network connection
var offlineMode = false
var useCached = ""

func TestCommandLineHelp(t *testing.T) {
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
		{"default", split("--help")},
	}
	run(t, tests)
}

func TestNegCommandLines(t *testing.T) {
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
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
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
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
	}

	run(t, tests)
}

func TestConfigUsers(t *testing.T) {
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
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
	tests := []struct {
		name string
		args struct{ args []string }
	}{
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
		{"neg-config-add-user-bad-use-encrypted",
			split("config add-user --username foobar --auth-type other --encrypt-password")},
	}

	run(t, tests)
}

func TestGetTags(t *testing.T) {
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
		{"get-tags",
			split("install mssql get-tags")},
	}

	run(t, tests)
}

func TestMssqlInstall(t *testing.T) {
	setup(t.Name())
	tests := []struct {
		name string
		args struct{ args []string }
	}{
		{"install",
			split(fmt.Sprintf("install mssql%v --user-database my-database --accept-eula --encrypt-password", useCached))},
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
	cmd := cmdparser.New[*Root](root.SubCommands()...)
	cmd.ArgsForUnitTesting(tt.args.args)

	viper.SetConfigFile(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd",
		"sqlconfig-"+t.Name(),
	))

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
		cmd.Execute()
	}
	cmd.Execute()
}

func Test_displayHints(t *testing.T) {
	displayHints([]string{"Test Hint"})
}

func TestIsValidRootCommand(t *testing.T) {
	IsValidSubCommand("install")
	IsValidSubCommand("create")
	IsValidSubCommand("nope")
}

func TestRunCommand(t *testing.T) {
	loggingLevel = 4
	Execute()
}

func Test_checkErr(t *testing.T) {
	loggingLevel = 3

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	checkErr(errors.New("Expected error"))
}

func run(t *testing.T, tests []struct {
	name string
	args struct{ args []string }
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runTests(t, tt) })
	}

	verifyConfigIsEmpty(t)
}

func verifyConfigIsEmpty(t *testing.T) {
	if !config.IsEmpty() {
		bytes := output.Struct(config.GetRedactedConfig(true))
		t.Error(fmt.Sprintf(
			"Config is not empty. Content of config file:\n%s\nConfig file used:%s",
			string(bytes),
			config.GetConfigFileUsed(),
		))
		t.Fail()
	}
}

func setup(testName string) {
	useCached = " --cached"
	if !offlineMode {
		useCached = ""
	}

	internal.Initialize(
		func(err error) {
			if err != nil {
				panic(err)
			}
		},
		displayHints,
		pal.FilenameInUserHomeDotDirectory(
			".sqlcmd",
			"sqlconfig-"+testName,
		),
		"yaml",
		4,
	)

	config.Clean()
}

type args struct {
	args []string
}

func split(cmd string) args {
	return args{strings.Split(cmd, " ")}
}
