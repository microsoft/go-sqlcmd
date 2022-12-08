// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		errorHandler func(error)
		hintHandler  func([]string)
		outputType   string
		loggingLevel int
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			func(err error) {
				if err != nil {
					panic(err)
				}
			},
			func(strings []string) {},
			"yaml",
			2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := InitializeOptions{
				ErrorHandler: tt.args.errorHandler,
				HintHandler:  tt.args.hintHandler,
				OutputType:   tt.args.outputType,
				LoggingLevel: tt.args.loggingLevel,
			}
			Initialize(options)
		})
	}
}
