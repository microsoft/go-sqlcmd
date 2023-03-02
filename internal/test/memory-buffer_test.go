// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryBuffer_Write(t *testing.T) {
	buffer := NewMemoryBuffer()
	_, err := buffer.Write([]byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, "hello world", buffer.String())
}

func TestMemoryBuffer_Close(t *testing.T) {
	buffer := NewMemoryBuffer()
	assert.NoError(t, buffer.Close())
}

func TestMemoryBuffer_String(t *testing.T) {
	buffer := NewMemoryBuffer()
	_, err := buffer.Write([]byte("foo bar"))
	if err != nil {
		return
	}
	assert.Equal(t, "foo bar", buffer.String())
}
