// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

import (
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestInitialize(t *testing.T) {
	output := output.New(output.Options{HintHandler: func(hints []string) {

	}, ErrorHandler: func(err error) {

	}})
	options := InitializeOptions{
		ErrorHandler: func(err error) {
			if err != nil {
				panic(err)
			}
		},
		HintHandler:  func(strings []string) {},
		TraceHandler: output.Tracef,
		LineBreak:    "\n",
	}
	Initialize(options)
}

func TestNegInitialize(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	options := InitializeOptions{
		ErrorHandler: nil,
	}
	Initialize(options)
}

func TestNegInitialize2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	options := InitializeOptions{
		ErrorHandler: func(err error) {},
	}
	Initialize(options)
}

func TestNegInitialize3(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	options := InitializeOptions{
		ErrorHandler: func(err error) {},
		TraceHandler: func(format string, a ...any) {},
	}
	Initialize(options)
}

func TestNegInitialize4(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	options := InitializeOptions{
		ErrorHandler: func(err error) {},
		TraceHandler: func(format string, a ...any) {},
		HintHandler:  func(strings []string) {},
	}
	Initialize(options)
}

func TestNegInitialize5(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	options := InitializeOptions{
		ErrorHandler: func(err error) {},
		TraceHandler: func(format string, a ...any) {},
		HintHandler:  func(strings []string) {},
		LineBreak:    "",
	}
	Initialize(options)
}
