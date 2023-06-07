package mssqlcontainer

import "github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"

type InitializeOptions struct {
	ErrorHandler func(error)
	TraceHandler func(format string, a ...any)
}

func Initialize(options InitializeOptions) {
	if options.ErrorHandler == nil {
		panic("ErrorHandler is nil")
	}
	if options.TraceHandler == nil {
		panic("TraceHandler is nil")
	}

	mechanism.Initialize(options.ErrorHandler, options.TraceHandler)
}
