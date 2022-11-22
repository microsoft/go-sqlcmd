// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package secret provide functions to encrypting and decrypting text such
// that the text can be persisted in a configuration file (xml / yaml / json etc.)
package secret

import (
	"encoding/base64"
)

// Encode optionally encrypts the plainText and always base64 encodes it
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

// Decode always base64 decodes the cipherText and optionally decrypts it
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
