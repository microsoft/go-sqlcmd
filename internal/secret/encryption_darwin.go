// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

var encryptionMethods = []string{"none"}

func encrypt(plainText string) (cipherText string) {

	//BUG(stuartpa): Encryption not yet implemented on macOS, will use the KeyChain
	cipherText = plainText

	return
}

func decrypt(cipherText string) (secret string) {
	secret = cipherText

	//BUG(stuartpa): Encryption not yet implemented on macOS, will use the KeyChain
	return
}
