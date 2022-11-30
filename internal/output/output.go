// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package Output manages outputting text to the user.
//
// Trace("Something very low level.") - not localized
// Debug("Useful debugging information.") - not localized
// Info("Something noteworthy happened!") - localized
// Warn("You should probably take a look at this.") - localized
// Error("Something failed but I'm not quitting.") - localized
// Fatal("Bye.") - localized
//
//	calls os.Exit(1) after logging
//
// Panic("I'm bailing.") - not localized
//
//	calls panic() after logging
package output

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output/formatter"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/pkg/errors"
	"io"
	"regexp"
	"strings"
)

func NewOutput(
	formatter formatter.Formatter,
	loggingLevel verbosity.Enum,
	standardWriterCloser io.WriteCloser,
) Output {
	fmt.Println("GOT HERE2")

	if formatter == nil {
		panic("Format must not be nil")
	}
	return Output{
		formatter:           formatter,
		loggingLevel:        loggingLevel,
		standardWriteCloser: standardWriterCloser,
	}
}

func (o Output) Debugf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Debug {
		format = o.ensureEol(format)
		o.printf("DEBUG: "+format, a...)
	}
}

func (o Output) Errorf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Error {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "ERROR: " + format
		}
		o.printf(format, a...)
	}
}

func (o Output) Fatal(a ...any) {
	o.fatal([]string{}, a...)
}
func (o Output) FatalErr(err error) {
	checkErr(err)
}

func (o Output) Fatalf(format string, a ...any) {
	o.fatalf([]string{}, format, a...)
}

func (o Output) FatalfErrorWithHints(err error, hints []string, format string, a ...any) {
	o.fatalf(hints, format, a...)
	checkErr(err)
}

func (o Output) FatalfWithHints(hints []string, format string, a ...any) {
	o.fatalf(hints, format, a...)
}

func (o Output) FatalfWithHintExamples(hintExamples [][]string, format string, a ...any) {
	err := errors.New(fmt.Sprintf(format, a...))
	o.displayHintExamples(hintExamples)
	checkErr(err)
}

func (o Output) FatalWithHints(hints []string, a ...any) {
	o.fatal(hints, a...)
}

func (o Output) Infof(format string, a ...any) {
	o.infofWithHints([]string{}, format, a...)
}

func (o Output) InfofWithHints(hints []string, format string, a ...any) {
	o.infofWithHints(hints, format, a...)
}

func (o Output) InfofWithHintExamples(hintExamples [][]string, format string, a ...any) {
	if o.loggingLevel >= verbosity.Info {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "INFO:  " + format
		}
		o.printf(format, a...)
		o.displayHintExamples(hintExamples)
	}
}

func (o Output) Panic(a ...any) {
	panic(a)
}

func (o Output) Panicf(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}

func (o Output) Struct(in interface{}) (bytes []byte) {
	if o.formatter == nil {
		panic("Formatter is nil")
	}
	bytes = o.formatter.Serialize(in)

	return
}

func (o Output) Tracef(format string, a ...any) {
	if o.loggingLevel >= verbosity.Trace {
		format = o.ensureEol(format)
		o.printf("TRACE: "+format, a...)
	}
}

func (o Output) Warnf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Warn {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "WARN:  " + format
		}
		o.printf(format, a...)
	}
}

func (o Output) displayHintExamples(hintExamples [][]string) {
	var hints []string

	maxLengthHintText := 0
	for _, hintExample := range hintExamples {
		if len(hintExample) != 2 {
			panic("Hintexample must be 2 elements, a description, and an example")
		}

		if len(hintExample[0]) > maxLengthHintText {
			maxLengthHintText = len(hintExample[0])
		}
	}

	for _, hintExample := range hintExamples {
		padLength := maxLengthHintText - len(hintExample[0])
		hints = append(hints, fmt.Sprintf(
			"%v: %v%s",
			hintExample[0],
			strings.Repeat(" ", padLength),
			hintExample[1],
		))
	}
	displayHints(hints)
}

func (o Output) ensureEol(format string) string {
	if len(format) >= len(sqlcmd.SqlcmdEol) {
		if !strings.HasSuffix(format, sqlcmd.SqlcmdEol) {
			format = format + sqlcmd.SqlcmdEol
		}
	} else {
		format = sqlcmd.SqlcmdEol
	}
	return format
}

func (o Output) fatal(hints []string, a ...any) {
	err := errors.New(fmt.Sprintf("%v", a...))
	displayHints(hints)
	checkErr(err)
}

func (o Output) fatalf(hints []string, format string, a ...any) {
	err := errors.New(fmt.Sprintf(format, a...))
	displayHints(hints)
	checkErr(err)
}

func (o Output) infofWithHints(hints []string, format string, a ...any) {
	if o.loggingLevel >= verbosity.Info {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "INFO:  " + format
		}
		o.printf(format, a...)
		displayHints(hints)
	}
}

func (o Output) maskSecrets(text string) string {

	// Mask password from T/SQL e.g. ALTER LOGIN [sa] WITH PASSWORD = N'foo';
	r := regexp.MustCompile(`(PASSWORD.*\s?=.*\s?N?')(.*)(')`)
	text = r.ReplaceAllString(text, "$1********$3")
	return text
}

func (o Output) printf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	text = o.maskSecrets(text)
	_, err := o.standardWriteCloser.Write([]byte(text))
	checkErr(err)
}
