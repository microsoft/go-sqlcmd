// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package secret provide functions to encrypting and decrypting text such
// that the text can be persisted in a configuration file (xml / yaml / json etc.)
package secret

import (
	"encoding/base64"
)

// Encode takes a plain text string and a boolean indicating whether or not to
// encrypt the plain text using a password, and returns the resulting cipher text.
// If the plain text is an empty string, this function will panic.
func Encode(plainText string, encryptPassword bool) (cipherText string) {
	if plainText == "" {
		panic("Cannot encode/encrypt an empty string")
	}

	if encryptPassword {
		cipherText = encrypt(plainText)
	} else {
		cipherText = plainText
	}

	cipherText = base64.StdEncoding.EncodeToString([]byte(cipherText))

	return
}

// Decode takes a cipher text and a boolean indicating whether or not to decrypt
// the cipher text using a password, and returns the resulting plain text.
// If the cipher text is an empty string, this function will panic.
func Decode(cipherText string, decryptPassword bool) (plainText string) {
	if cipherText == "" {
		panic("Cannot decode/decrypt an empty string")
	}

	bytes, err := base64.StdEncoding.DecodeString(cipherText)
	checkErr(err)

	if decryptPassword {
		plainText = decrypt(string(bytes))
	} else {
		plainText = string(bytes)
	}

	return
}
