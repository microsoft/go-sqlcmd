// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		errorHandler   func(err error)
		traceHandler   func(format string, a ...any)
		hintHandler    func(hints []string)
		standardOutput io.WriteCloser
		errorOutput    io.WriteCloser
		format         string
		verbosity      verbosity.Enum
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "badFormatterPanic",
			args: args{
				errorHandler:   errorCallback,
				traceHandler:   traceCallback,
				hintHandler:    hintCallback,
				standardOutput: os.Stdout,
				errorOutput:    os.Stderr,
				format:         "badbad",
				verbosity:      0,
			},
		},
		{
			name: "initWithXml",
			args: args{
				errorHandler:   errorCallback,
				traceHandler:   traceCallback,
				hintHandler:    hintCallback,
				standardOutput: os.Stdout,
				errorOutput:    os.Stderr,
				format:         "xml",
				verbosity:      0,
			},
		},
		{
			name: "initWithJson",
			args: args{
				errorHandler:   errorCallback,
				traceHandler:   traceCallback,
				hintHandler:    hintCallback,
				standardOutput: os.Stdout,
				errorOutput:    os.Stderr,
				format:         "json",
				verbosity:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// If test name ends in 'Panic' expect a Panic
			if strings.HasSuffix(tt.name, "Panic") {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}
			Initialize(
				tt.args.errorHandler,
				tt.args.traceHandler,
				tt.args.hintHandler,
				tt.args.standardOutput,
				tt.args.format,
				tt.args.verbosity,
			)
		})
	}
}
