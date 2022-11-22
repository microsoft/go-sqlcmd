// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

import (
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		errorHandler      func(error)
		hintHandler       func([]string)
		sqlconfigFilename string
		outputType        string
		loggingLevel      int
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
			nil,
			pal.FilenameInUserHomeDotDirectory(
				".sqlcmd",
				"sqlconfig-test"),
			"yaml",
			0,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize(
				tt.args.errorHandler,
				tt.args.hintHandler,
				tt.args.sqlconfigFilename,
				tt.args.outputType,
				tt.args.loggingLevel,
			)
		})
	}
}
