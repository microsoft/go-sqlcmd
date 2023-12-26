// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package main is the entrypoint for sqlcmd. This package first initializes
// a new instance of the Root cmd then checks if the new cobra-based
// command-line interface (CLI) should be used based on if the first argument provided
// by the user is a valid sub-command for the new CLI, if so it executes the
// new cobra CLI; otherwise, it falls back to the old kong-based CLI.

//go:generate go-winres make --file-version=git-tag --product-version=git-tag

package main

import (
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/spf13/cobra"
	"path"

	"os"

	legacyCmd "github.com/microsoft/go-sqlcmd/cmd/sqlcmd"
)

var rootCmd *Root
var outputter *output.Output
var version = "local-build" // overridden in pipeline builds with: -ldflags="-X main.version=$(VersionTag)"

// main is the entry point for the sqlcmd command line interface.
// It parses command line options and initializes the command parser.
// If the first argument is a modern CLI subcommand, the modern CLI is
// executed. Otherwise, the legacy CLI is executed.
func main() {
	dependencies := dependency.Options{
		Output: output.New(output.Options{
			StandardWriter: os.Stdout,
			ErrorHandler:   checkErr,
			HintHandler:    displayHints})}
	rootCmd = cmdparser.New[*Root](dependencies)
	if isFirstArgModernCliSubCommand() {
		cmdparser.Initialize(initializeCallback)
		rootCmd.Execute()
	} else {
		initializeEnvVars()
		legacyCmd.Execute(version)
	}
}

// initializeEnvVars intializes SQLCMDSERVER, SQLCMDUSER and SQLCMDPASSWORD
// if the currentContext is set and if these env vars are not already set.
// In terms of precedence, command line switches/flags take higher precedence
// than env variables and env variables take higher precedence over config
// file info.
func initializeEnvVars() {
	home, err := os.UserHomeDir()

	// Special case, some shells don't have any home dir env var set, see:
	//   https://github.com/microsoft/go-sqlcmd/issues/279
	//  in this case, we early exit here and don't initialize anything, there is nothing
	//  else we can do
	if err != nil {
		return
	}

	// The only place we can check for the existence of the sqlconfig file is in the
	// default location, because this code path is only used for the legacy kong CLI,
	// if the sqlconfig file doesn't exist at the default location we just return so
	// because the initializeCallback() function below will create the sqlconfig file,
	// which legacy sqlcmd users might not want.
	if !file.Exists(path.Join(home, ".sqlcmd", "sqlconfig")) {
		return
	}

	initializeCallback()
	if config.CurrentContextName() != "" {
		server, username, password := config.GetCurrentContextInfo()
		if os.Getenv("SQLCMDSERVER") == "" {
			os.Setenv("SQLCMDSERVER", server)
		}

		// Username and password should come together, either from the environment
		// variables set by the user before the sqlcmd process started, or from the sqlconfig
		// for the current context, but if just the environment variable SQLCMDPASSWORD
		// is set before the process starts we do not use it, if the user and password is set in sqlconfig.
		if username != "" && password != "" { // If not trusted auth
			if os.Getenv("SQLCMDUSER") == "" {
				os.Setenv("SQLCMDUSER", username)
				os.Setenv("SQLCMDPASSWORD", password)
			}
		}
	}
}

// isFirstArgModernCliSubCommand is TEMPORARY code, to be removed when
// we remove the Kong based CLI
func isFirstArgModernCliSubCommand() (isNewCliCommand bool) {
	if len(os.Args) > 1 {
		if rootCmd.IsValidSubCommand(os.Args[1]) {
			isNewCliCommand = true
		}
	}
	return
}

// initializeCallback is called after the command line has been parsed and
// all values provided by the user are now available
func initializeCallback() {

	// Assigns a new outputter now that we have the outputType and loggingLevel
	// provided to us from the user
	outputter = output.New(
		output.Options{
			StandardWriter: os.Stdout,
			ErrorHandler:   checkErr,
			HintHandler:    displayHints,
			OutputType:     rootCmd.outputType,
			LoggingLevel:   verbosity.Level(rootCmd.loggingLevel),
		})
	internal.Initialize(
		internal.InitializeOptions{
			ErrorHandler: checkErr,
			TraceHandler: outputter.Tracef,
			HintHandler:  displayHints,
			LineBreak:    sqlcmd.SqlcmdEol,
			LoggingLevel: verbosity.Level(rootCmd.loggingLevel),
		})
	mssqlcontainer.Initialize(mssqlcontainer.InitializeOptions{
		ErrorHandler: checkErr,
		TraceHandler: outputter.Tracef,
	})
	config.SetFileName(rootCmd.configFilename)
	config.Load()
}

// checkErr uses Cobra to check err, and halts the application if err is not
// nil.  Pass (inject) checkErr into all dependencies (internal helpers etc.) as an
// errorHandler.
//
// To aid debugging issues, if the logging level is > 2 (e.g. --verbosity 3 or --verbosity 4), we
// panic which outputs a stacktrace.
func checkErr(err error) {
	if rootCmd != nil && rootCmd.loggingLevel > 2 {
		if err != nil {
			panic(err)
		}
	} else {
		cobra.CheckErr(err)
	}
}

// displayHints displays helpful information on what the user should do next
// to make progress.  displayHints is injected into dependencies (helpers etc.)
func displayHints(hints []string) {
	if len(hints) > 0 {
		outputter.Infof("%vHINT:", pal.LineBreak())
		for i, hint := range hints {
			outputter.Infof("  %d. %v", i+1, hint)
		}
		outputter.Infof("")
	}
}
