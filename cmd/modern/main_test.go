// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"errors"
	"github.com/microsoft/go-sqlcmd/internal/buffer"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMainStart(t *testing.T) {
	os.Args[1] = "--help"
	main()
}

func TestInitializeCallback(t *testing.T) {
	rootCmd = cmdparser.New[*Root](dependency.Options{})
	initializeCallback()
}

func TestDisplayHints(t *testing.T) {
	buf := buffer.NewMemoryBuffer()
	outputter = output.New(output.Options{StandardWriter: buf})
	displayHints([]string{"This is a hint"})
	assert.Equal(t, "\r"+pal.LineBreak()+
		"HINT:"+
		pal.LineBreak()+
		"  1. This is a hint"+pal.LineBreak()+pal.LineBreak(), buf.String())
	err := buf.Close()
	checkErr(err)
}

func TestCheckErr(t *testing.T) {
	rootCmd = cmdparser.New[*Root](dependency.Options{})
	rootCmd.loggingLevel = 4
	checkErr(nil)
	assert.Panics(t, func() {
		checkErr(errors.New("test error"))
	})
}
