// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"errors"
	"github.com/microsoft/go-sqlcmd/internal/output/formatter"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"os"
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
			loggingLevel := tt.args.loggingLevel
			o := new(loggingLevel)
			o.Tracef(tt.args.format, tt.args.a...)
			o.Debugf(tt.args.format, tt.args.a...)
			o.Infof(tt.args.format, tt.args.a...)
			o.Warnf(tt.args.format, tt.args.a...)
			o.Errorf(tt.args.format, tt.args.a...)
			o.Struct(tt.args.a)

			o.InfofWithHints([]string{}, tt.args.format, tt.args.a...)
			o.InfofWithHintExamples([][]string{}, tt.args.format, tt.args.a...)
		})
	}
}

func new(loggingLevel verbosity.Enum) Output {
	return NewOutput(
		formatter.NewFormatter("yaml", os.Stdout, func(err error) {
			if err != nil {
				panic(err)
			}
		}),
		loggingLevel,
		os.Stdout,
	)
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
			o := new(4)
			o.Fatal(tt.args.a...)
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
			o := new(4)
			o.FatalWithHints(tt.args.hints, tt.args.a...)
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
			o := new(4)
			o.FatalfWithHintExamples(tt.args.hintExamples, tt.args.format, tt.args.a...)
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
			o := new(4)
			o.FatalfErrorWithHints(tt.args.err, tt.args.hints, tt.args.format, tt.args.a...)
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
			o := new(4)
			o.FatalfWithHints(tt.args.hints, tt.args.format, tt.args.a...)
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
			o := new(4)
			o.Fatalf(tt.args.format, tt.args.a...)
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
			o := new(4)
			o.FatalErr(tt.args.err)
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
			o := new(4)
			o.Panicf(tt.args.format, tt.args.a...)
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
			o := new(4)
			o.Panic(tt.args.a...)
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

			o := new(4)
			o.InfofWithHintExamples(tt.args.hintExamples, tt.args.format, tt.args.a...)
		})
	}
}
