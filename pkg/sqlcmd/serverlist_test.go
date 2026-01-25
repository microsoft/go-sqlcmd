// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListLocalServers(t *testing.T) {
	// Test that ListLocalServers writes to the provided writer without error
	// Note: actual server discovery depends on SQL Browser service availability
	var buf bytes.Buffer
	ListLocalServers(&buf)
	// We can't assert specific content since it depends on environment,
	// but we verify it doesn't panic and writes valid output
	t.Logf("ListLocalServers output: %q", buf.String())
}

func TestGetLocalServerInstances(t *testing.T) {
	// Test that GetLocalServerInstances returns a slice (may be empty if no servers)
	instances, err := GetLocalServerInstances()
	// instances may be nil or empty if no SQL Browser is running, that's OK
	// err may be non-nil for non-timeout network errors
	if err != nil {
		t.Logf("GetLocalServerInstances returned error (expected in some environments): %v", err)
	}
	t.Logf("Found %d instances", len(instances))
	for _, inst := range instances {
		assert.NotEmpty(t, inst, "Instance name should not be empty")
	}
}

func TestParseInstances(t *testing.T) {
	// Test parsing of SQL Browser response
	// Format: 0x05 (response type), 2 bytes length, then semicolon-separated key=value pairs
	// Each instance ends with two semicolons

	t.Run("empty response", func(t *testing.T) {
		result := parseInstances([]byte{})
		assert.Empty(t, result)
	})

	t.Run("invalid header", func(t *testing.T) {
		result := parseInstances([]byte{1, 0, 0})
		assert.Empty(t, result)
	})

	t.Run("valid single instance", func(t *testing.T) {
		// Simulating SQL Browser response format
		// Header: 0x05 followed by 2 length bytes, then the instance data
		data := []byte{5, 0, 0}
		instanceData := "ServerName;MYSERVER;InstanceName;MSSQLSERVER;IsClustered;No;Version;15.0.2000.5;tcp;1433;;"
		data = append(data, []byte(instanceData)...)

		result := parseInstances(data)
		assert.Len(t, result, 1)
		assert.Contains(t, result, "MSSQLSERVER")
		assert.Equal(t, "MYSERVER", result["MSSQLSERVER"]["ServerName"])
		assert.Equal(t, "1433", result["MSSQLSERVER"]["tcp"])
	})

	t.Run("valid multiple instances", func(t *testing.T) {
		data := []byte{5, 0, 0}
		instanceData := "ServerName;MYSERVER;InstanceName;MSSQLSERVER;tcp;1433;;ServerName;MYSERVER;InstanceName;SQLEXPRESS;tcp;1434;;"
		data = append(data, []byte(instanceData)...)

		result := parseInstances(data)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "MSSQLSERVER")
		assert.Contains(t, result, "SQLEXPRESS")
	})
}

func TestServerlistCommand(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()

	// Run the serverlist command
	c := []string{":serverlist"}
	err := runSqlCmd(t, s, c)

	// The command should not raise an error even if no servers are found
	assert.NoError(t, err, ":serverlist should not raise error")
	// Output may be empty if no SQL Browser is running
	t.Logf("Serverlist output: %q", buf.buf.String())
}
