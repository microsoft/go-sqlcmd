// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	notEncrypted := Encode("plainText", false)
	encrypted := Encode("plainText", true)
	Decode(notEncrypted, false)
	Decode(encrypted, true)
}

func TestNegEncode(t *testing.T) {
	assert.Panics(t, func() {

		Encode("", true)
	})
}

func TestNegDecode(t *testing.T) {
	assert.Panics(t, func() {

		Decode("", true)
	})
}

func TestDecodeAsUtf16(t *testing.T) {
	cipherText := Encode("plainText", true)
	DecodeAsUtf16(cipherText, true)
}
