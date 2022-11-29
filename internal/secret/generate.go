// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"crypto/rand"
	"math/big"
	mathRand "math/rand"
	"strings"
)

const (
	lowerCharSet = "abcdedfghijklmnopqrstuvwxyz"
	upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberSet    = "0123456789"
)

func Generate(passwordLength, minSpecialChar, minNum, minUpperCase int, specialCharSet string) string {
	var password strings.Builder
	allCharSet := lowerCharSet + upperCharSet + specialCharSet + numberSet

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
		checkErr(err)
		_, err = password.WriteString(string(specialCharSet[idx.Int64()]))
		checkErr(err)
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))
		checkErr(err)
		_, err = password.WriteString(string(numberSet[idx.Int64()]))
		checkErr(err)
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))
		checkErr(err)
		_, err = password.WriteString(string(upperCharSet[idx.Int64()]))
		checkErr(err)
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
		checkErr(err)
		_, err = password.WriteString(string(allCharSet[idx.Int64()]))
		checkErr(err)
	}

	inRune := []rune(password.String())
	mathRand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}
