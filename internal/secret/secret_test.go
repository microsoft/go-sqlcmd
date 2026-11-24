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

func TestNegEncodeBadEncryptionMethod(t *testing.T) {
	assert.Panics(t, func() {
		Encode("plainText", "does-not-exist")
	})
}

func TestNegDecode(t *testing.T) {
	assert.Panics(t, func() {
		Decode("", "none")
	})
}

func TestNegDecodeBadEncryptionMethod(t *testing.T) {
	assert.Panics(t, func() {
		Decode("cipherText", "does-not-exist")
	})
}

// TestEncryptionMethodsForUsage ensures at least "none" is an available
// encryption method for usage.
func TestEncryptionMethodsForUsage(t *testing.T) {
	s := EncryptionMethodsForUsage()
	assert.Contains(t, s, "none")
}

func TestIsValidEncryptionMethod(t *testing.T) {
	assert.True(t, IsValidEncryptionMethod("none"))
	assert.False(t, IsValidEncryptionMethod("does-not-exist"))
}
