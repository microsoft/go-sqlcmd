// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	notEncrypted := Encode("plainText", "none")
	encrypted := Encode("plainText", "dpapi")
	Decode(notEncrypted, "none")
	Decode(encrypted, "dpapi")
}

func TestNegEncode(t *testing.T) {
	assert.Panics(t, func() {
		Encode("", "dpapi")
	})
}

func TestNegDecode(t *testing.T) {
	assert.Panics(t, func() {
		Decode("", "dpapi")
	})
}

func TestDecodeAsUtf16(t *testing.T) {
	cipherText := Encode("plainText", "dpapi")
	DecodeAsUtf16(cipherText, "dpapi")
}
