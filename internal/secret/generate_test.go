// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerate(t *testing.T) {
	type args struct {
		passwordLength int
		minSpecialChar int
		minNum         int
		minUpperCase   int
		specialChars   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positiveSimple",
			args: args{
				passwordLength: 50,
				minSpecialChar: 10,
				minNum:         10,
				minUpperCase:   10,
				specialChars:   "!@#$%&*",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(
				tt.args.passwordLength,
				tt.args.minSpecialChar,
				tt.args.minNum,
				tt.args.minUpperCase,
				tt.args.specialChars,
			)
			assert.Len(t, got, tt.args.passwordLength)
		})
	}
}
