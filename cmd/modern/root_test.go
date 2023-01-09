// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRoot is a quick sanity test
func TestRoot(t *testing.T) {
	c := cmdparser.New[*Root](dependency.Options{})
	c.DefineCommand()
	c.SetArgsForUnitTesting([]string{})
	c.Execute()
}

func TestIsValidSubCommand(t *testing.T) {
	c := cmdparser.New[*Root](dependency.Options{})
	invalid := c.IsValidSubCommand("nope")
	assert.Equal(t, false, invalid)
	valid := c.IsValidSubCommand("query")
	assert.Equal(t, true, valid)
}
