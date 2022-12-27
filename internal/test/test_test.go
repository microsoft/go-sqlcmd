// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"errors"
	"testing"
)

func TestCatchExpectedError(t *testing.T) {
	CatchExpectedError(errors.New("test"), t)
}

func TestCatchExpectedError2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	CatchExpectedError(nil, t)
}
