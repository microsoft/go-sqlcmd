// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/billgraziano/dpapi"
)

var encryptionMethods = []string{"none", "dpapi"}

func encrypt(plainText string) (cipherText string) {
	var err error

	cipherText, err = dpapi.Encrypt(plainText)
	checkErr(err)

	return
}

func decrypt(cipherText string) (secret string) {
	var err error

	secret, err = dpapi.Decrypt(cipherText)
	checkErr(err)

	return
}
