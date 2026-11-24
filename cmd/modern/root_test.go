// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"bytes"
	"testing"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/stretchr/testify/assert"
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
	assert.False(t, c.IsValidSubCommand("open"))
}

func TestRootHelpDoesNotAdvertiseADS(t *testing.T) {
	c := cmdparser.New[*Root](dependency.Options{})
	var output bytes.Buffer
	c.Command().SetOut(&output)
	c.SetArgsForUnitTesting([]string{"--help"})

	err := c.Command().Execute()

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "sqlcmd query")
	assert.NotContains(t, output.String(), "sqlcmd open")
	assert.NotContains(t, output.String(), "Azure Data Studio")
}

func TestRootRejectsOpenCommand(t *testing.T) {
	for _, args := range [][]string{{"open"}, {"open", "ads"}} {
		t.Run(args[len(args)-1], func(t *testing.T) {
			c := cmdparser.New[*Root](dependency.Options{})
			c.Command().SetOut(&bytes.Buffer{})
			c.Command().SetErr(&bytes.Buffer{})
			c.SetArgsForUnitTesting(args)

			err := c.Command().Execute()

			assert.Error(t, err)
		})
	}
}
