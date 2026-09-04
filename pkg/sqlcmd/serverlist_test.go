// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"
	"testing"

	"github.com/microsoft/go-mssqldb/msdsn"
	"github.com/stretchr/testify/assert"
)

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestListLocalServers(t *testing.T) {
	original := getLocalServerInstances
	getLocalServerInstances = func() ([]string, error) {
		return []string{`MYSERVER\SQL2019`, `MYSERVER\SQL2022`}, nil
	}
	defer func() { getLocalServerInstances = original }()

	var buf bytes.Buffer

	assert.NoError(t, ListLocalServers(&buf))
	assert.Equal(t, "  MYSERVER\\SQL2019"+SqlcmdEol+"  MYSERVER\\SQL2022"+SqlcmdEol, buf.String())

	writeErr := errors.New("write failed")
	assert.ErrorIs(t, ListLocalServers(failingWriter{err: writeErr}), writeErr)
}

func TestParseInstances(t *testing.T) {
	// Test parsing of SQL Browser response
	// Format: 0x05 (response type), 2 bytes length, then alternating key;value tokens
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

	t.Run("missing final terminator", func(t *testing.T) {
		data := append([]byte{5, 0, 0}, []byte("ServerName;MYSERVER;InstanceName;SQLEXPRESS;tcp;1434")...)

		result := parseInstances(data)

		assert.Equal(t, "MYSERVER", result["SQLEXPRESS"]["ServerName"])
		assert.Equal(t, "1434", result["SQLEXPRESS"]["tcp"])
	})
}

func TestLocalServerInstanceNamesSkipsMissingServerNames(t *testing.T) {
	data := msdsn.BrowserData{
		"MISSING": {"InstanceName": "MISSING"},
		"EMPTY":   {"ServerName": "", "InstanceName": "EMPTY"},
		"VALID":   {"ServerName": "MYSERVER", "InstanceName": "VALID"},
	}

	assert.Equal(t, []string{`MYSERVER\VALID`}, localServerInstanceNames(data))
}

func TestIsBrowserUnavailableError(t *testing.T) {
	assert.True(t, isBrowserUnavailableError(fmt.Errorf("browser unavailable: %w", syscall.ECONNREFUSED)))
	assert.True(t, isBrowserUnavailableError(fmt.Errorf("browser unavailable: %w", syscall.ECONNRESET)))
	assert.False(t, isBrowserUnavailableError(errors.New("network failure")))
}

func TestServerlistCommand(t *testing.T) {
	original := getLocalServerInstances
	getLocalServerInstances = func() ([]string, error) {
		return []string{`MYSERVER\SQL2019`}, nil
	}
	defer func() { getLocalServerInstances = original }()

	v := InitializeVariables(false)
	s := New(nil, "", v)
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetOutput(buf)
	defer func() { _ = buf.Close() }()

	err := serverlistCommand(s, []string{""}, 1)

	assert.NoError(t, err)
	assert.Equal(t, "  MYSERVER\\SQL2019"+SqlcmdEol, buf.buf.String())
}

func TestServerlistCommandWritesErrors(t *testing.T) {
	discoveryErr := errors.New("network failure")
	original := getLocalServerInstances
	getLocalServerInstances = func() ([]string, error) {
		return nil, discoveryErr
	}
	defer func() { getLocalServerInstances = original }()

	s := New(nil, "", InitializeVariables(false))
	errBuf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetError(errBuf)
	defer func() { _ = errBuf.Close() }()

	assert.NoError(t, serverlistCommand(s, []string{""}, 1))
	assert.Equal(t, discoveryErr.Error()+SqlcmdEol, errBuf.buf.String())
}
