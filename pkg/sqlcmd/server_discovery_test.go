// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInstanceData(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[string]map[string]string
	}{
		{
			name: "single default instance",
			// Format: 0x05 (response type), 2 bytes length, then key;value;key;value;;
			input: []byte{0x05, 0x00, 0x00,
				'S', 'e', 'r', 'v', 'e', 'r', 'N', 'a', 'm', 'e', ';', 'M', 'Y', 'S', 'E', 'R', 'V', 'E', 'R', ';',
				'I', 'n', 's', 't', 'a', 'n', 'c', 'e', 'N', 'a', 'm', 'e', ';', 'M', 'S', 'S', 'Q', 'L', 'S', 'E', 'R', 'V', 'E', 'R', ';',
				'I', 's', 'C', 'l', 'u', 's', 't', 'e', 'r', 'e', 'd', ';', 'N', 'o', ';',
				'V', 'e', 'r', 's', 'i', 'o', 'n', ';', '1', '5', '.', '0', '.', '2', '0', '0', '0', '.', '5', ';',
				't', 'c', 'p', ';', '1', '4', '3', '3', ';',
				';'},
			expected: map[string]map[string]string{
				"MSSQLSERVER": {
					"ServerName":   "MYSERVER",
					"InstanceName": "MSSQLSERVER",
					"IsClustered":  "No",
					"Version":      "15.0.2000.5",
					"tcp":          "1433",
				},
			},
		},
		{
			name:     "empty response - too short",
			input:    []byte{0x05, 0x00},
			expected: map[string]map[string]string{},
		},
		{
			name:     "empty response - wrong type",
			input:    []byte{0x04, 0x00, 0x00, 'a', ';', 'b', ';', ';'},
			expected: map[string]map[string]string{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: map[string]map[string]string{},
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: map[string]map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseInstanceData(tc.input)
			assert.Equal(t, len(tc.expected), len(result), "Number of instances should match")
			for instName, expectedProps := range tc.expected {
				actualProps, ok := result[instName]
				assert.True(t, ok, "Instance %s should exist", instName)
				for key, expectedValue := range expectedProps {
					assert.Equal(t, expectedValue, actualProps[key], "Property %s should match", key)
				}
			}
		})
	}
}

func TestFormatServerList(t *testing.T) {
	tests := []struct {
		name      string
		instances []ServerInstance
		expected  []string
	}{
		{
			name:      "empty list",
			instances: []ServerInstance{},
			expected:  []string{},
		},
		{
			name: "default instance only",
			instances: []ServerInstance{
				{ServerName: "MYSERVER", InstanceName: "MSSQLSERVER", IsClustered: "No", Version: "15.0", Port: "1433"},
			},
			expected: []string{"(local)", "MYSERVER"},
		},
		{
			name: "named instance only",
			instances: []ServerInstance{
				{ServerName: "MYSERVER", InstanceName: "SQL2019", IsClustered: "No", Version: "15.0", Port: "1434"},
			},
			expected: []string{"MYSERVER\\SQL2019"},
		},
		{
			name: "multiple instances mixed",
			instances: []ServerInstance{
				{ServerName: "SERVER1", InstanceName: "MSSQLSERVER", IsClustered: "No", Version: "15.0", Port: "1433"},
				{ServerName: "SERVER1", InstanceName: "DEV", IsClustered: "No", Version: "14.0", Port: "1435"},
				{ServerName: "SERVER2", InstanceName: "PROD", IsClustered: "Yes", Version: "15.0", Port: "1436"},
			},
			expected: []string{"(local)", "SERVER1", "SERVER1\\DEV", "SERVER2\\PROD"},
		},
		{
			name: "named instance with different case preserved",
			instances: []ServerInstance{
				{ServerName: "MyServer", InstanceName: "SqlExpress", IsClustered: "No", Version: "15.0", Port: "1433"},
			},
			expected: []string{"MyServer\\SqlExpress"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatServerList(tc.instances)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatServerListPreservesOrder(t *testing.T) {
	// Verify that the order of instances is preserved in output
	instances := []ServerInstance{
		{ServerName: "ALPHA", InstanceName: "MSSQLSERVER"},
		{ServerName: "BETA", InstanceName: "TEST"},
	}

	result := FormatServerList(instances)

	// Default instance should produce (local) then ServerName
	assert.Equal(t, "(local)", result[0])
	assert.Equal(t, "ALPHA", result[1])
	// Named instance should be ServerName\InstanceName
	assert.Equal(t, "BETA\\TEST", result[2])
}
