// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"github.com/microsoft/go-sqlcmd/cmd/root"
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

var loggingLevel int
var outputType string
var configFilename string
var rootCmd cmdparser.Command

// Initialize initializes the command-line interface. The func passed into
// cmdparser.Initialize is called after the command-line from the user has been
// parsed, so the helpers are initialized with the values from the command-line
// like '-v 4' which sets the logging level to maximum etc.
func Initialize() {
	cmdparser.Initialize(initialize)
	rootCmd = cmdparser.New[*Root](root.SubCommands()...)
}

func initialize() {
	options := internal.InitializeOptions{
		ErrorHandler: checkErr,
		HintHandler:  displayHints,
		OutputType:   "yaml",
		LoggingLevel: 2,
	}

	config.SetFileName(configFilename)
	config.Load()
	internal.Initialize(options)
}

// Execute runs the application based on the command-line
// parameters the user has passed in.
func Execute() {
	rootCmd.Execute()
}

// IsValidSubCommand is TEMPORARY code, that will be removed when
// we enable the new cobra based CLI by default.  It returns true if the
// command-line provided by the user indicates they want the new cobra
// based CLI, e.g. sqlcmd install, or sqlcmd query, or sqlcmd --help etc.
func IsValidSubCommand(command string) bool {
	return rootCmd.IsSubCommand(command)
}

// checkErr uses Cobra to check err, and halts the application if err is not
// nil.  Pass (inject) checkErr into all dependencies (helpers etc.) as an
// errorHandler.
//
// To aid debugging issues, if the logging level is > 2 (e.g. -v 3 or -4), we
// panic which outputs a stacktrace.
func checkErr(err error) {
	if loggingLevel > 2 {
		if err != nil {
			panic(err)
		}
	}
	rootCmd.CheckErr(err)
}

// displayHints displays helpful information on what the user should do next
// to make progress.  displayHints is injected into dependencies (helpers etc.)
func displayHints(hints []string) {
	if len(hints) > 0 {
		output.Infof("\nHINT:")
		for i, hint := range hints {
			output.Infof("  %d. %v", i+1, hint)
		}
	}
}
