// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestFactory(t *testing.T) {
	o := New(Options{unitTesting: false, HintHandler: func(hints []string) {

	}, ErrorHandler: func(err error) {

	}})
	o.errorCallback(nil)
}

func TestNegtactory(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	New(Options{unitTesting: true,
		HintHandler:  func(hints []string) {},
		ErrorHandler: nil})
}

func TestNegFactory2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	New(Options{unitTesting: true,
		HintHandler:  nil,
		ErrorHandler: func(err error) {}})
}
