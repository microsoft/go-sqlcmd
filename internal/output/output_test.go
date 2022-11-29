// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"errors"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"testing"
)

func TestTracef(t *testing.T) {
	type args struct {
		loggingLevel verbosity.Enum
		format       string
		a            []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			loggingLevel: verbosity.Trace,
			format:       "%v",
			a:            []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggingLevel = tt.args.loggingLevel
			Tracef(tt.args.format, tt.args.a...)
			Debugf(tt.args.format, tt.args.a...)
			Infof(tt.args.format, tt.args.a...)
			Warnf(tt.args.format, tt.args.a...)
			Errorf(tt.args.format, tt.args.a...)
			Struct(tt.args.a)

			InfofWithHints([]string{}, tt.args.format, tt.args.a...)
			InfofWithHintExamples([][]string{}, tt.args.format, tt.args.a...)
		})
	}
}

func TestFatal(t *testing.T) {
	type args struct {
		a []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			a: []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			Fatal(tt.args.a...)
		})
	}
}

func TestFatalWithHints(t *testing.T) {
	type args struct {
		hints []string
		a     []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			hints: []string{"This is a hint"},
			a:     []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			FatalWithHints(tt.args.hints, tt.args.a...)
		})
	}
}

func TestFatalfWithHintExamples(t *testing.T) {
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
			hintExamples: [][]string{{"This is a hint", "With a sample"}},
			a:            []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			FatalfWithHintExamples(tt.args.hintExamples, tt.args.format, tt.args.a...)
		})
	}
}

func TestFatalfErrorWithHints(t *testing.T) {
	type args struct {
		err    error
		hints  []string
		format string
		a      []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			hints: []string{"This is a hint"},
			a:     []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			FatalfErrorWithHints(tt.args.err, tt.args.hints, tt.args.format, tt.args.a...)
		})
	}
}

func TestFatalfWithHints(t *testing.T) {
	type args struct {
		hints  []string
		format string
		a      []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			hints: []string{"This is a hint"},
			a:     []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			FatalfWithHints(tt.args.hints, tt.args.format, tt.args.a...)
		})
	}
}

func TestFatalf(t *testing.T) {
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			a: []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			Fatalf(tt.args.format, tt.args.a...)
		})
	}
}

func TestFatalErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			errors.New("an error"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			FatalErr(tt.args.err)
		})
	}
}

func TestPanicf(t *testing.T) {
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			a: []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			Panicf(tt.args.format, tt.args.a...)
		})
	}
}

func TestPanic(t *testing.T) {
	type args struct {
		a []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"default", args{
			a: []any{"sample trace"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			Panic(tt.args.a...)
		})
	}
}

func TestInfofWithHintExamples(t *testing.T) {
	t.Skip() // BUG(stuartpa): CrossPlatScripts build is failing on this test!?  (presume this is an issue with static state, move to an object)
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
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()

			//BUG(stuartpa): Not thread safe
			runningUnitTests = true
			InfofWithHintExamples(tt.args.hintExamples, tt.args.format, tt.args.a...)
			runningUnitTests = false
		})
	}
}

func Test_ensureEol(t *testing.T) {
	format := ensureEol("%s")
	Infof(format, "hello-world")
}
