// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package secret

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	notEncrypted := Encode("plainText", false)
	encrypted := Encode("plainText", true)
	Decode(notEncrypted, false)
	Decode(encrypted, true)
}

func TestNegEncode(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	Encode("", true)
}

func TestNegDecode(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	Decode("", true)
}
