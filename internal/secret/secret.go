// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package secret provide functions to encrypting and decrypting text such
// that the text can be persisted in a configuration file (xml / yaml / json etc.)
package secret

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf16"
)

// Encode takes a plain text string and a boolean indicating whether or not to
// encrypt the plain text using a password, and returns the resulting cipher text.
// If the plain text is an empty string, this function will panic.
func Encode(plainText string, passwordEncryption string) (cipherText string) {
	if plainText == "" {
		panic("Cannot encode/encrypt an empty string")
	}
	if !IsValidEncryptionMethod(passwordEncryption) {
		panic(fmt.Sprintf(
			"Invalid encryption method (%q not in %q)",
			passwordEncryption,
			EncryptionMethodsForUsage(),
		))
	}

	if passwordEncryption != "none" {
		cipherText = encrypt(plainText)
	} else {
		cipherText = plainText
	}

	cipherText = base64.StdEncoding.EncodeToString([]byte(cipherText))

	return
}

// Decode takes a cipher text and a boolean indicating whether to decrypt
// the cipher text using a password, and returns the resulting plain text.
// If the cipher text is an empty string, this function will panic.
func Decode(cipherText string, passwordEncryption string) (plainText string) {
	if cipherText == "" {
		panic("Cannot decode/decrypt an empty string")
	}

	// BUG(stuartpa): Temporary code, remove when rolling over to v1.0.0, needed
	// as very early users migrated from the old password-encrypted bool value
	if passwordEncryption == "" {
		passwordEncryption = "none"
	}

	if !IsValidEncryptionMethod(passwordEncryption) {
		panic(fmt.Sprintf(
			"Invalid encryption method (%q not in %q)",
			passwordEncryption,
			EncryptionMethodsForUsage(),
		))
	}

	bytes, err := base64.StdEncoding.DecodeString(cipherText)
	checkErr(err)

	if passwordEncryption != "none" {
		plainText = decrypt(string(bytes))
	} else {
		plainText = string(bytes)
	}

	return
}

// DecodeAsUtf16 takes a cipher text and a boolean indicating whether to decrypt
// and returns the resulting plain text as a byte array in UTF-16 format which
// is required when passing the secret to applications written using managed
// code (C#), such as Azure Data Studio.
func DecodeAsUtf16(cipherText string, passwordEncryption string) []byte {
	var buf bytes.Buffer
	secret := Decode(cipherText, passwordEncryption)
	runes := utf16.Encode([]rune(secret))

	var b [2]byte
	for _, r := range runes {
		b[1] = byte(r >> 8)
		b[0] = byte(r & 255)
		buf.Write(b[:])
	}

	return buf.Bytes()
}

// EncryptionMethodsForUsage return the encryption methods for the current
// OS as comma seperated string for use in help text
func EncryptionMethodsForUsage() string {
	return strings.Join(encryptionMethods, ", ")
}

// IsValidEncryptionMethod returns true if the method is a valid encryption method
func IsValidEncryptionMethod(method string) bool {
	for _, m := range encryptionMethods {
		if m == method {
			return true
		}
	}
	return false
}
