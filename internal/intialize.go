// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

import (
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/http"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/net"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
)

type InitializeOptions struct {
	ErrorHandler func(error)
	TraceHandler func(format string, a ...any)
	HintHandler  func([]string)
	LineBreak    string
	LoggingLevel verbosity.Level
}

// Initialize initializes various dependencies for the application with the provided options.
// The dependencies that are initialized include file, sql, config, container,
// secret, net, and pal. This function is typically called at the start of the application
// to ensure that all dependencies are properly initialized before any other code is executed.
func Initialize(options InitializeOptions) {
	if options.ErrorHandler == nil {
		panic("ErrorHandler is nil")
	}
	if options.TraceHandler == nil {
		panic("TraceHandler is nil")
	}
	if options.HintHandler == nil {
		panic("HintHandler is nil")
	}
	if options.LineBreak == "" {
		panic("LineBreak is empty")
	}

	enableTraceLogging := false
	if options.LoggingLevel == verbosity.Trace {
		enableTraceLogging = true
	}
	file.Initialize(options.ErrorHandler, options.TraceHandler)
	sql.Initialize(enableTraceLogging, options.ErrorHandler, options.TraceHandler, secret.Decode)
	config.Initialize(options.ErrorHandler, options.TraceHandler, secret.Encode, secret.Decode, net.IsLocalPortAvailable)
	container.Initialize(options.ErrorHandler, options.TraceHandler)
	secret.Initialize(options.ErrorHandler)
	net.Initialize(options.ErrorHandler, options.TraceHandler)
	http.Initialize(options.ErrorHandler, options.TraceHandler)
	pal.Initialize(options.ErrorHandler, options.LineBreak)
}
