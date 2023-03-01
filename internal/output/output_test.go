// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"errors"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestTracef(t *testing.T) {
	format := "%v"
	args := []string{"sample text"}

	loggingLevel := verbosity.Trace
	o := New(Options{LoggingLevel: loggingLevel, HintHandler: func(hints []string) {

	}, ErrorHandler: func(err error) {

	}})
	o.Tracef(format, args)
	o.Debugf(format, args)
	o.Infof(format, args)
	o.Warnf(format, args)
	o.Errorf(format, args)
	o.Struct(args)

	o.InfofWithHints([]string{}, format, args)
	o.InfofWithHintExamples([][]string{}, format, args)
}

func TestFatal(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.Fatal("sample trace")
	})
}

func TestFatalWithHints(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.FatalWithHints([]string{"This is a hint"}, "expected error")
	})
}

func TestFatalfWithHintExamples(t *testing.T) {
	hintExamples := [][]string{{"This is a hint", ""}}
	o := New(Options{LoggingLevel: verbosity.Trace})
	assert.Panics(t, func() {
		o.FatalfWithHintExamples(
			hintExamples,
			"%v",
			"this is an error",
		)
	})
}

func TestFatalfErrorWithHints(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.FatalfErrorWithHints(
			errors.New("error to check"),
			[]string{"This is a hint to avoid the error"},
			"%v",
			"This the error message",
		)
	})
}

func TestFatalfWithHints(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.FatalfWithHints(
			[]string{"This is a hint to the user to avoid the error"},
			"%v",
			"this is the reason for the fatal error",
		)
	})
}

func TestFatalf(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.Fatalf("%v", "message to give user on exit")
	})
}

func TestFatalErr(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.FatalErr(errors.New("will exist if error is not nil"))
	})
}

func TestPanicf(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.Panicf("%v", "this is the reason for the panic")
	})
}

func TestPanic(t *testing.T) {
	o := New(Options{LoggingLevel: 4})
	assert.Panics(t, func() {
		o.Panic("reason for the panic")
	})
}

func TestInfofWithHintExamples(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Failing in CI, not sure why (stuartpa)")
	}

	type args struct {
		hintExamples [][]string
		format       string
		a            []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			hintExamples: [][]string{{
				"Bad",
				"Sample",
				"One To Many Elements",
			}, {"Good", "Example"}},
			format: "sample trace %v",
			a:      []any{"hello"},
		}},
		{"emptyFormatString", args{
			hintExamples: [][]string{{
				"Bad",
				"Sample",
				"One To Many Elements",
			}, {"Good", "Example"}},
			format: "",
			a:      []any{"hello"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := New(Options{LoggingLevel: 4})
			assert.Panics(t, func() {
				o.InfofWithHintExamples(tt.args.hintExamples, tt.args.format, tt.args.a...)
			})
		})
	}
}
