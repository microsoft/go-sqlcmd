// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

func encrypt(plainText string) (cipherText string) {

	//BUG(stuartpa): Encryption not yet implemented on linux
	cipherText = plainText

	return
}

func decrypt(cipherText string) (secret string) {
	secret = cipherText

	//BUG(stuartpa): Encryption not yet implemented on linux

	return
}
