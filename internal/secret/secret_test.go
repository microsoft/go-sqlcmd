// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	notEncrypted := Encode("plainText", "none")
	plainText := Decode(notEncrypted, "none")
	assert.Equal(t, "plainText", plainText)

	if runtime.GOOS == "windows" {
		encrypted := Encode("plainText", "dpapi")
		plainText := Decode(encrypted, "dpapi")
		assert.Equal(t, "plainText", plainText)
	}
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
	plainText := DecodeAsUtf16(cipherText, "dpapi")
	assert.Equal(t, "plainText", plainText)
}
