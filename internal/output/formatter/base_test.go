// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"strings"
	"testing"
)

func TestBase_CheckErr(t *testing.T) {
	type fields struct {
		ErrorHandlerCallback func(err error)
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "noCallBackHandlerPanic",
			fields: fields{ErrorHandlerCallback: nil},
			args:   args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Base{
				ErrorHandlerCallback: tt.fields.ErrorHandlerCallback,
			}

			// If test name ends in 'Panic' expect a Panic
			if strings.HasSuffix(tt.name, "Panic") {
				defer func() { test.CatchExpectedError(recover(), t) }()
			}

			f.CheckErr(tt.args.err)
		})
	}
}
