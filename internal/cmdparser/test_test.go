// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

// Test_test.go contains functions to test the functions in test.go, so this file
// tests the test functions.  This file shows end-to-end usage of how to create
// the simplest command-line application and run it

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestCommand struct {
	Cmd

	throwError string
}

func (c *TestCommand) DefineCommand(...CommandOptions) {
	options := CommandOptions{}
	options.Use = "test-cmd"
	options.Short = "A test command"
	options.FirstArgAlternativeForFlag = &AlternativeForFlagOptions{
		Flag:  "throw-error",
		Value: &c.throwError,
	}
	options.Run = func() {
		c.Output().InfofWithHints([]string{"This is a hint"}, "Some things to consider")

		if c.throwError == "throw-error" {
			c.CheckErr(errors.New("Expected error"))
		}
	}

	c.Cmd.DefineCommand(options)
	c.AddFlag(FlagOptions{Name: "throw-error", Usage: "Throw an error", String: &c.throwError})
}

func TestTest(t *testing.T) {
	TestSetup(t)
	TestCmd[*TestCommand]()
}

func TestTest2(t *testing.T) {
	TestSetup(t)
	TestCmd[*TestCommand]("test-cmd")
}

func TestThrowError(t *testing.T) {
	assert.Panics(t, func() {
		TestSetup(t)
		TestCmd[*TestCommand]("throw-error")
	})
}

func TestTest3(t *testing.T) {
	TestSetup(t)
}
