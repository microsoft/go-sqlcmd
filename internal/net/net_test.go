// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package net

import (
	"testing"
)

func TestIsLocalPortAvailable(t *testing.T) {
	t.Skip() // BUG(stuartpa): Re-enable before merge, fix to work on any machine
	type args struct {
		port int
	}
	tests := []struct {
		name              string
		args              args
		wantPortAvailable bool
	}{
		{name: "expectedToNotBeAvailable", args: args{port: 51027}, wantPortAvailable: false},
		{name: "expectedToBeAvailable", args: args{port: 9999}, wantPortAvailable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPortAvailable := IsLocalPortAvailable(tt.args.port); gotPortAvailable != tt.wantPortAvailable {
				t.Errorf("IsLocalPortAvailable() = %v, want %v", gotPortAvailable, tt.wantPortAvailable)
			}
		})
	}
}
