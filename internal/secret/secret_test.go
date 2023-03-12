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
		Encode("", "none")
	})
}

func TestNegDecode(t *testing.T) {
	assert.Panics(t, func() {
		Decode("", "none")
	})
}

func TestDecodeAsUtf16(t *testing.T) {
	cipherText := Encode("plainText", "none")
	plainText := DecodeAsUtf16(cipherText, "none")
	assert.Equal(t, []byte{0x70, 0x0, 0x6c, 0x0, 0x61, 0x0, 0x69, 0x0, 0x6e, 0x0, 0x54, 0x0, 0x65, 0x0, 0x78, 0x0, 0x74, 0x0}, plainText)
}
