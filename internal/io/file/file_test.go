// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file_test

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ExampleCreateEmptyIfNotExists() {
	filename := filepath.Join(os.TempDir(), "foo.txt")

	file.CreateEmptyIfNotExists(filename)
}

func TestFileExamples(t *testing.T) {
	ExampleCreateEmptyIfNotExists()
}

func TestCreateEmptyIfNotExists(t *testing.T) {
	filename := "foo.txt"
	folderName := "folder"

	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "default", args: args{filename: filename}},
		{name: "alreadyExists", args: args{filename: filename}},
		{name: "emptyInputPanic", args: args{filename: ""}},
		{name: "incFolder", args: args{filename: filepath.Join(folderName, filename)}},
	}

	cleanup(folderName, filename)
	defer cleanup(folderName, filename)

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

			file.CreateEmptyIfNotExists(tt.args.filename)
		})
	}
}

func TestExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
	}{
		{name: "exists", args: args{filename: "file_test.go"}, wantExists: true},
		{name: "notExists", args: args{filename: "does-not-exist.file"}, wantExists: false},
		{name: "noFilenamePanic", args: args{filename: ""}},
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

			if gotExists := file.Exists(tt.args.filename); gotExists != tt.wantExists {
				t.Errorf("Exists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func cleanup(folderName string, filename string) {
	if file.Exists(folderName) {
		folder.RemoveAll(folderName)
	}

	if file.Exists(filename) {
		file.Remove(filename)
	}
}

func TestCloseFile(t *testing.T) {
	f := file.OpenFile("test.txt")
	file.CloseFile(f)
}

func TestGetContents(t *testing.T) {
	f := file.OpenFile("test.txt")
	defer file.CloseFile(f)
	file.WriteString(f, "test contents")
	contents := file.GetContents("test.txt")
	if contents != "test contents" {
		t.Errorf("Expected contents to be 'test contents', but got '%s'", contents)
	}
}

func TestGetContentsBadFileName(t *testing.T) {
	assert.Panics(t, func() {
		file.GetContents("badbad.txt")
	})
}

func TestOpenFile(t *testing.T) {
	f := file.OpenFile("test.txt")
	_, err := os.Stat("test.txt")
	if err != nil {
		t.Error("Expected file to be created, but it does not exist")
	}
	file.CloseFile(f)
	file.Remove("test.txt")
}

func TestWriteString(t *testing.T) {
	f := file.OpenFile("test.txt")
	file.WriteString(f, "test contents")
	contents := file.GetContents("test.txt")
	if contents != "test contents" {
		t.Errorf("Expected contents to be 'test contents', but got '%s'", contents)
	}
	file.CloseFile(f)
	file.Remove("test.txt")
}
