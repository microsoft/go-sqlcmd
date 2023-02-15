// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestFormatter(t *testing.T) {
	s := []string{"serialize this"}

	var f Formatter
	f = New(Options{SerializationFormat: "yaml"})
	f.Serialize(s)
	f = New(Options{SerializationFormat: "xml"})
	f.Serialize(s)
	f = New(Options{SerializationFormat: "json"})
	f.Serialize(s)

	log.Println("This is here to ensure a newline is in test output")
}

func TestNegFormatterBadFormat(t *testing.T) {
	assert.Panics(t, func() {

		s := "serialize this"
		f := New(Options{SerializationFormat: "badbad"})
		f.Serialize(s)
	})
}

func TestFormatterEmptyFormat(t *testing.T) {
	s := "serialize this"
	f := New(Options{SerializationFormat: ""})
	f.Serialize(s)
}
