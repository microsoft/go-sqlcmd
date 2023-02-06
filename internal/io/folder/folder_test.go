// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package folder

import (
	"os"
	"strings"
	"testing"
)

func TestMkdirAll(t *testing.T) {
	folderName := "test-folder"
	type args struct {
		folder string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "default", args: args{folder: folderName}},
		{name: "noFolderNamePanic", args: args{folder: ""}},
	}

	cleanup(folderName)
	defer cleanup(folderName)

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

			MkdirAll(tt.args.folder)
		})
	}
}

func TestGetwd(t *testing.T) {
	// Test 1: Check that the function returns the correct path
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got := Getwd(); got != wd {
		t.Errorf("Getwd() = %q, want %q", got, wd)
	}
}

func cleanup(folderName string) {
	RemoveAll(folderName)
}
